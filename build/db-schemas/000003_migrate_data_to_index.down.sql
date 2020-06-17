INSERT INTO cells_identification_service_areas
SELECT UNNEST(cells) as cell_id,
       13 AS cell_level,
       id AS identification_service_area_id
FROM identification_service_areas
ON CONFLICT (identification_service_area_id, cell_id)
DO NOTHING;

INSERT INTO cells_subscriptions
SELECT UNNEST(cells) AS cell_id,
       13 AS cell_level,
       id AS subscription_id
FROM subscriptions
ON CONFLICT (subscription_id, cell_id)
DO NOTHING;