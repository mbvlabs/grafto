-- River migration 003 [down]
ALTER TABLE river_job ALTER COLUMN tags DROP NOT NULL,
                      ALTER COLUMN tags DROP DEFAULT;
