-- Test data for FinanKal

-- Insert test users
INSERT INTO users (id, name, username, created_at) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'John Doe', 'johndoe', NOW()),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Jane Smith', 'janesmith', NOW());

-- Insert test accounts linked to users
INSERT INTO accounts (id, user_id, name, type, created_at) VALUES
('11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Checking Account', 'ASSET', NOW()),
('22222222-2222-2222-2222-222222222222', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Savings Account', 'ASSET', NOW()),
('33333333-3333-3333-3333-333333333333', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Business Account', 'ASSET', NOW());

-- Insert test transactions
INSERT INTO transactions (id, user_id, description, created_at) VALUES
('44444444-4444-4444-4444-444444444444', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Initial deposit', NOW()),
('55555555-5555-5555-5555-555555555555', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Transfer between accounts', NOW());

-- Insert test entries
INSERT INTO entries (transaction_id, account_id, amount, type, created_at) VALUES
('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 500.00, 'DEBIT', NOW()),
('44444444-4444-4444-4444-444444444444', '22222222-2222-2222-2222-222222222222', 500.00, 'CREDIT', NOW()),
('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', 100.00, 'DEBIT', NOW()),
('55555555-5555-5555-5555-555555555555', '22222222-2222-2222-2222-222222222222', 100.00, 'CREDIT', NOW());
