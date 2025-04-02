INSERT INTO institutions (name)
VALUES ('rk');
INSERT INTO institution_email_masks (email_mask, institution_id)
VALUES ('teachmask.edu', (SELECT id FROM institutions WHERE name = 'rk' LIMIT 1))