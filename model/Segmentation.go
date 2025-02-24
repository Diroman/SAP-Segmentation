package model

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"sap_segmentation/config"
)

type Row struct {
	ID           int64  `db:"id" json:"id"`
	AddressSapID string `db:"address_sap_id" json:"address_sap_id"`
	AdrSegment   string `db:"adr_segment" json:"adr_segment"`
	SegmentID    int64  `db:"segment_id" json:"segment_id"`
}

type Segmentation struct {
	db         *sqlx.DB
	cfg        *config.Config
	logger     *logrus.Logger
	client     *http.Client
	retryCount int
}

func NewSegmentation(cfg *config.Config, logger *logrus.Logger) (*Segmentation, error) {
	db, err := connectDB(cfg)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: time.Duration(cfg.ConnTimeout) * time.Second,
	}

	return &Segmentation{
		db:         db,
		cfg:        cfg,
		logger:     logger,
		client:     client,
		retryCount: 3,
	}, nil
}

func connectDB(cfg *config.Config) (*sqlx.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPassword)
	return sqlx.Connect("postgres", connStr)
}

func (s *Segmentation) Run(_ context.Context) error {
	offset := 0

	for {
		req, err := s.buildRequest(offset)
		if err != nil {
			return errors.Wrap(err, "build request")
		}

		s.logger.Printf("Requesting data from: %s", req.URL.String())

		rows, err := s.doRequestRetry(req)
		if err != nil {
			return errors.Wrap(err, "request error")
		}
		if len(rows) == 0 {
			break
		}

		if err = s.saveRows(rows); err != nil {
			return errors.Wrap(err, "save rows")
		}

		offset += s.cfg.ImportBatchSize
		time.Sleep(time.Duration(s.cfg.ConnInterval) * time.Millisecond)
	}

	s.logger.Info("No more data to import")
	return nil
}

func (s *Segmentation) buildRequest(offset int) (*http.Request, error) {
	baseURL, err := url.Parse(s.cfg.ConnURI)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse connection uri")
	}

	params := url.Values{}
	params.Add("p_limit", strconv.Itoa(s.cfg.ImportBatchSize))
	params.Add("p_offset", strconv.Itoa(offset))
	baseURL.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("User-Agent", s.cfg.ConnUserAgent)
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", s.cfg.ConnAuthLoginPwd))

	return req, nil
}

func (s *Segmentation) doRequestRetry(req *http.Request) ([]*Row, error) {
	for i := 0; i < s.retryCount; i++ {
		rows, err := s.doRequest(req)
		if err != nil {
			s.logger.WithError(err).WithField("attempt", i).Errorf("request error")
			time.Sleep(time.Duration(s.cfg.ConnInterval) * time.Millisecond)
			continue
		}

		return rows, nil
	}

	return nil, errors.New("max retries exceeded")
}

func (s *Segmentation) doRequest(req *http.Request) ([]*Row, error) {
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute HTTP request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if len(body) == 0 {
		return nil, nil
	}

	s.logger.Info(string(body))

	var rows []*Row
	if err = json.Unmarshal(body, &rows); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON response")
	}

	return rows, nil
}

func (s *Segmentation) saveRows(rows []*Row) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	data, _ := json.MarshalIndent(rows, "", "  ")
	s.logger.Info(string(data))

	query := `
        INSERT INTO segmentation (address_sap_id, adr_segment, segment_id)
        VALUES (:address_sap_id, :adr_segment, :segment_id)
        ON CONFLICT (address_sap_id) 
        DO UPDATE SET 
            adr_segment = EXCLUDED.adr_segment,
            segment_id = EXCLUDED.segment_id`

	for _, segment := range rows {
		if _, err = tx.NamedExec(query, segment); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
