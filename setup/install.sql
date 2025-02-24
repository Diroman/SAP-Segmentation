CREATE TABLE IF NOT EXISTS segmentation (
    id BIGSERIAL PRIMARY KEY,
    address_sap_id VARCHAR(255) NOT NULL,
    adr_segment VARCHAR(16),
    segment_id BIGINT,
    CONSTRAINT unique_address_sap_id UNIQUE (address_sap_id)
);