-- name: RecordEconomicSnapshot :one
INSERT INTO economic_snapshots (economic_day, metrics)
VALUES ($1, $2)
RETURNING *;

-- name: CreateAnomalyFlag :one
INSERT INTO anomaly_flags (flag_type, implicated_player_ids, evidence_summary)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetOpenAnomalies :many
SELECT * FROM anomaly_flags
WHERE review_status = 'OPEN'
ORDER BY detected_at DESC;

-- name: UpdateAnomalyStatus :exec
UPDATE anomaly_flags SET review_status = $1
WHERE id = $2;

-- name: LogAdminAction :one
INSERT INTO admin_audit_log (admin_player_id, target_player_id, action_type, parameters)
VALUES ($1, $2, $3, $4)
RETURNING *;