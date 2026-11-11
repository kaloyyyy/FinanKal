-- Keep the occurrence status constraint in sync with the lifecycle used by the engine.
BEGIN;

ALTER TABLE scheduled_transaction_occurrences
    DROP CONSTRAINT IF EXISTS chk_scheduled_transaction_occurrence_status;

ALTER TABLE scheduled_transaction_occurrences
    ADD CONSTRAINT chk_scheduled_transaction_occurrence_status
        CHECK (status IN ('PENDING', 'GENERATED', 'COMPLETED', 'SKIPPED', 'FAILED'));

COMMIT;
