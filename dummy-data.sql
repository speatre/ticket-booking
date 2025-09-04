-- ===========================================
-- TICKET BOOKING SYSTEM - DUMMY DATA
-- ===========================================
-- This script populates the database with realistic dummy data
-- for development and testing purposes
--
-- Run this after running the migration:
-- psql -U postgres -d ticket_booking < dummy-data.sql
-- ===========================================

-- ===========================================
-- USERS DATA
-- ===========================================

-- Admin users (password: admin123)
INSERT INTO users (id, email, password_hash, role, full_name, created_at, updated_at) VALUES
(uuid_generate_v4(), 'admin@ticketbooking.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'ADMIN', 'System Administrator', NOW() - INTERVAL '30 days', NOW()),
(uuid_generate_v4(), 'manager@ticketbooking.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'ADMIN', 'Event Manager', NOW() - INTERVAL '25 days', NOW());

-- Regular users (password: password123)
INSERT INTO users (id, email, password_hash, role, full_name, created_at, updated_at) VALUES
(uuid_generate_v4(), 'john.doe@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'USER', 'John Doe', NOW() - INTERVAL '20 days', NOW()),
(uuid_generate_v4(), 'jane.smith@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'USER', 'Jane Smith', NOW() - INTERVAL '18 days', NOW()),
(uuid_generate_v4(), 'mike.johnson@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'USER', 'Mike Johnson', NOW() - INTERVAL '15 days', NOW()),
(uuid_generate_v4(), 'sarah.wilson@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'USER', 'Sarah Wilson', NOW() - INTERVAL '12 days', NOW()),
(uuid_generate_v4(), 'david.brown@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'USER', 'David Brown', NOW() - INTERVAL '10 days', NOW()),
(uuid_generate_v4(), 'lisa.davis@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'USER', 'Lisa Davis', NOW() - INTERVAL '8 days', NOW()),
(uuid_generate_v4(), 'alex.garcia@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'USER', 'Alex Garcia', NOW() - INTERVAL '6 days', NOW()),
(uuid_generate_v4(), 'emma.martinez@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'USER', 'Emma Martinez', NOW() - INTERVAL '4 days', NOW()),
(uuid_generate_v4(), 'ryan.anderson@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'USER', 'Ryan Anderson', NOW() - INTERVAL '2 days', NOW()),
(uuid_generate_v4(), 'olivia.taylor@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'USER', 'Olivia Taylor', NOW() - INTERVAL '1 day', NOW());

-- ===========================================
-- EVENTS DATA
-- ===========================================

-- Past Events (completed)
INSERT INTO events (id, name, description, starts_at, ends_at, capacity, remaining, ticket_price_cents, created_at, updated_at) VALUES
(uuid_generate_v4(), 'Coldplay World Tour', 'Experience Coldplay''s spectacular live performance with hits spanning two decades', '2024-01-15 19:30:00+00', '2024-01-15 23:00:00+00', 50000, 0, 12500, NOW() - INTERVAL '60 days', NOW() - INTERVAL '14 days'),
(uuid_generate_v4(), 'NBA Finals Game 1', 'Los Angeles Lakers vs Boston Celtics - Championship basketball at its finest', '2024-02-01 20:00:00+00', '2024-02-01 22:30:00+00', 20000, 0, 20000, NOW() - INTERVAL '55 days', NOW() - INTERVAL '13 days'),
(uuid_generate_v4(), 'Hamilton Musical', 'Award-winning Broadway musical about America''s Founding Father', '2024-02-20 20:00:00+00', '2024-02-20 22:30:00+00', 1500, 0, 18000, NOW() - INTERVAL '50 days', NOW() - INTERVAL '12 days'),
(uuid_generate_v4(), 'Taylor Swift Eras Tour', 'The tour that celebrates the artist''s entire career in one spectacular show', '2024-03-10 19:00:00+00', '2024-03-10 23:00:00+00', 70000, 0, 15000, NOW() - INTERVAL '45 days', NOW() - INTERVAL '11 days'),
(uuid_generate_v4(), 'Champions League Final', 'The biggest night in European club football featuring top teams', '2024-03-25 21:00:00+00', '2024-03-25 23:30:00+00', 80000, 0, 25000, NOW() - INTERVAL '40 days', NOW() - INTERVAL '10 days');

-- Current/Future Events (available)
INSERT INTO events (id, name, description, starts_at, ends_at, capacity, remaining, ticket_price_cents, created_at, updated_at) VALUES
(uuid_generate_v4(), 'Summer Music Festival 2024', 'Three days of non-stop music featuring 50+ artists across multiple stages', '2024-07-15 12:00:00+00', '2024-07-17 23:00:00+00', 75000, 25000, 9500, NOW() - INTERVAL '30 days', NOW()),
(uuid_generate_v4(), 'Ed Sheeran Mathematics Tour', 'Intimate acoustic performance in a stunning venue setting', '2024-08-05 20:00:00+00', '2024-08-05 22:30:00+00', 12000, 4500, 8500, NOW() - INTERVAL '25 days', NOW()),
(uuid_generate_v4(), 'Wimbledon Tennis Championship', 'The Championships - one of the four Grand Slam tournaments', '2024-07-01 11:00:00+00', '2024-07-14 19:00:00+00', 40000, 15000, 12000, NOW() - INTERVAL '20 days', NOW()),
(uuid_generate_v4(), 'Broadway on Tour - The Lion King', 'Disney''s award-winning musical spectacular with stunning costumes and effects', '2024-09-15 19:30:00+00', '2024-09-15 22:00:00+00', 2500, 800, 22000, NOW() - INTERVAL '15 days', NOW()),
(uuid_generate_v4(), 'Formula 1 Grand Prix', 'High-speed racing action with the world''s best drivers and teams', '2024-08-25 14:00:00+00', '2024-08-25 16:00:00+00', 100000, 35000, 7500, NOW() - INTERVAL '10 days', NOW()),
(uuid_generate_v4(), 'Jazz in the Park', 'Free outdoor jazz concert featuring local and international artists', '2024-06-20 18:00:00+00', '2024-06-20 21:00:00+00', 5000, 2000, 0, NOW() - INTERVAL '5 days', NOW()),
(uuid_generate_v4(), 'Comedy Night Special', 'Stand-up comedy featuring top comedians from around the world', '2024-08-30 20:00:00+00', '2024-08-30 22:00:00+00', 800, 320, 4500, NOW() - INTERVAL '3 days', NOW()),
(uuid_generate_v4(), 'Classical Music Symphony', 'Full orchestra performance of Beethoven''s greatest works', '2024-09-08 19:00:00+00', '2024-09-08 21:30:00+00', 2000, 950, 3500, NOW() - INTERVAL '2 days', NOW());

-- Special/Premium Events
INSERT INTO events (id, name, description, starts_at, ends_at, capacity, remaining, ticket_price_cents, created_at, updated_at) VALUES
(uuid_generate_v4(), 'Metallica World Tour 2024', 'Heavy metal legends bring their signature high-energy performance', '2024-10-15 18:30:00+00', '2024-10-15 22:30:00+00', 40000, 5000, 11000, NOW() - INTERVAL '35 days', NOW()),
(uuid_generate_v4(), 'VIP Broadway Experience', 'Exclusive behind-the-scenes tour plus premium seating for Chicago', '2024-11-05 16:00:00+00', '2024-11-05 22:00:00+00', 200, 45, 50000, NOW() - INTERVAL '28 days', NOW()),
(uuid_generate_v4(), 'Super Bowl LVIII', 'The biggest game of the year featuring top NFL teams', '2024-02-11 18:30:00+00', '2024-02-11 22:30:00+00', 75000, 0, 8000, NOW() - INTERVAL '75 days', NOW() - INTERVAL '16 days'),
(uuid_generate_v4(), 'New Year''s Eve Celebration', 'Spectacular fireworks and live entertainment to ring in the new year', '2024-12-31 20:00:00+00', '2025-01-01 01:00:00+00', 30000, 12000, 6500, NOW() - INTERVAL '1 day', NOW());

-- ===========================================
-- BOOKINGS DATA
-- ===========================================

-- Function to get user IDs for bookings
CREATE OR REPLACE FUNCTION get_user_id_by_email(user_email TEXT)
RETURNS UUID AS $$
BEGIN
    RETURN (SELECT id FROM users WHERE email = user_email);
END;
$$ LANGUAGE plpgsql;

-- Function to get event IDs for bookings
CREATE OR REPLACE FUNCTION get_event_id_by_name(event_name TEXT)
RETURNS UUID AS $$
BEGIN
    RETURN (SELECT id FROM events WHERE name = event_name);
END;
$$ LANGUAGE plpgsql;

-- Bookings for past events (completed)
INSERT INTO bookings (id, user_id, event_id, quantity, status, created_at, updated_at) VALUES
(uuid_generate_v4(), get_user_id_by_email('john.doe@example.com'), get_event_id_by_name('Coldplay World Tour'), 2, 'CONFIRMED', NOW() - INTERVAL '50 days', NOW() - INTERVAL '14 days'),
(uuid_generate_v4(), get_user_id_by_email('jane.smith@example.com'), get_event_id_by_name('Coldplay World Tour'), 4, 'CONFIRMED', NOW() - INTERVAL '48 days', NOW() - INTERVAL '14 days'),
(uuid_generate_v4(), get_user_id_by_email('mike.johnson@example.com'), get_event_id_by_name('NBA Finals Game 1'), 1, 'CONFIRMED', NOW() - INTERVAL '45 days', NOW() - INTERVAL '13 days'),
(uuid_generate_v4(), get_user_id_by_email('sarah.wilson@example.com'), get_event_id_by_name('Hamilton Musical'), 2, 'CONFIRMED', NOW() - INTERVAL '42 days', NOW() - INTERVAL '12 days'),
(uuid_generate_v4(), get_user_id_by_email('david.brown@example.com'), get_event_id_by_name('Taylor Swift Eras Tour'), 3, 'CONFIRMED', NOW() - INTERVAL '40 days', NOW() - INTERVAL '11 days'),
(uuid_generate_v4(), get_user_id_by_email('lisa.davis@example.com'), get_event_id_by_name('Champions League Final'), 1, 'CONFIRMED', NOW() - INTERVAL '38 days', NOW() - INTERVAL '10 days'),
(uuid_generate_v4(), get_user_id_by_email('alex.garcia@example.com'), get_event_id_by_name('Super Bowl LVIII'), 2, 'CONFIRMED', NOW() - INTERVAL '70 days', NOW() - INTERVAL '16 days'),
(uuid_generate_v4(), get_user_id_by_email('emma.martinez@example.com'), get_event_id_by_name('Super Bowl LVIII'), 1, 'CONFIRMED', NOW() - INTERVAL '68 days', NOW() - INTERVAL '16 days');

-- Current bookings (active)
INSERT INTO bookings (id, user_id, event_id, quantity, status, created_at, updated_at) VALUES
(uuid_generate_v4(), get_user_id_by_email('john.doe@example.com'), get_event_id_by_name('Summer Music Festival 2024'), 3, 'CONFIRMED', NOW() - INTERVAL '20 days', NOW() - INTERVAL '1 day'),
(uuid_generate_v4(), get_user_id_by_email('jane.smith@example.com'), get_event_id_by_name('Ed Sheeran Mathematics Tour'), 2, 'CONFIRMED', NOW() - INTERVAL '18 days', NOW() - INTERVAL '2 days'),
(uuid_generate_v4(), get_user_id_by_email('mike.johnson@example.com'), get_event_id_by_name('Wimbledon Tennis Championship'), 1, 'CONFIRMED', NOW() - INTERVAL '15 days', NOW() - INTERVAL '3 days'),
(uuid_generate_v4(), get_user_id_by_email('sarah.wilson@example.com'), get_event_id_by_name('Broadway on Tour - The Lion King'), 4, 'PENDING', NOW() - INTERVAL '12 days', NOW() - INTERVAL '1 day'),
(uuid_generate_v4(), get_user_id_by_email('david.brown@example.com'), get_event_id_by_name('Formula 1 Grand Prix'), 2, 'CONFIRMED', NOW() - INTERVAL '8 days', NOW() - INTERVAL '2 days'),
(uuid_generate_v4(), get_user_id_by_email('lisa.davis@example.com'), get_event_id_by_name('Jazz in the Park'), 1, 'CONFIRMED', NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 day'),
(uuid_generate_v4(), get_user_id_by_email('alex.garcia@example.com'), get_event_id_by_name('Comedy Night Special'), 3, 'PENDING', NOW() - INTERVAL '2 days', NOW() - INTERVAL '30 minutes'),
(uuid_generate_v4(), get_user_id_by_email('emma.martinez@example.com'), get_event_id_by_name('Classical Music Symphony'), 2, 'CONFIRMED', NOW() - INTERVAL '1 day', NOW() - INTERVAL '2 hours');

-- Pending bookings (awaiting payment)
INSERT INTO bookings (id, user_id, event_id, quantity, status, created_at, updated_at) VALUES
(uuid_generate_v4(), get_user_id_by_email('ryan.anderson@example.com'), get_event_id_by_name('Summer Music Festival 2024'), 2, 'PENDING', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '30 minutes'),
(uuid_generate_v4(), get_user_id_by_email('olivia.taylor@example.com'), get_event_id_by_name('Metallica World Tour 2024'), 1, 'PENDING', NOW() - INTERVAL '45 minutes', NOW() - INTERVAL '20 minutes'),
(uuid_generate_v4(), get_user_id_by_email('john.doe@example.com'), get_event_id_by_name('VIP Broadway Experience'), 1, 'PENDING', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '1 hour');

-- Cancelled bookings
INSERT INTO bookings (id, user_id, event_id, quantity, status, created_at, updated_at) VALUES
(uuid_generate_v4(), get_user_id_by_email('mike.johnson@example.com'), get_event_id_by_name('Taylor Swift Eras Tour'), 2, 'CANCELLED', NOW() - INTERVAL '35 days', NOW() - INTERVAL '11 days'),
(uuid_generate_v4(), get_user_id_by_email('sarah.wilson@example.com'), get_event_id_by_name('Champions League Final'), 1, 'CANCELLED', NOW() - INTERVAL '32 days', NOW() - INTERVAL '10 days'),
(uuid_generate_v4(), get_user_id_by_email('david.brown@example.com'), get_event_id_by_name('Summer Music Festival 2024'), 1, 'CANCELLED', NOW() - INTERVAL '10 days', NOW() - INTERVAL '5 days');

-- Premium bookings
INSERT INTO bookings (id, user_id, event_id, quantity, status, created_at, updated_at) VALUES
(uuid_generate_v4(), get_user_id_by_email('admin@ticketbooking.com'), get_event_id_by_name('VIP Broadway Experience'), 2, 'CONFIRMED', NOW() - INTERVAL '25 days', NOW() - INTERVAL '2 days'),
(uuid_generate_v4(), get_user_id_by_email('manager@ticketbooking.com'), get_event_id_by_name('New Year''s Eve Celebration'), 5, 'CONFIRMED', NOW() - INTERVAL '20 days', NOW() - INTERVAL '1 day');

-- ===========================================
-- UPDATE REMAINING TICKETS BASED ON BOOKINGS
-- ===========================================

-- Update remaining tickets for each event based on confirmed bookings
UPDATE events SET remaining = capacity - (
    SELECT COALESCE(SUM(quantity), 0)
    FROM bookings
    WHERE bookings.event_id = events.id
    AND bookings.status = 'CONFIRMED'
);

-- ===========================================
-- CLEANUP FUNCTIONS
-- ===========================================

DROP FUNCTION IF EXISTS get_user_id_by_email(TEXT);
DROP FUNCTION IF EXISTS get_event_id_by_name(TEXT);

-- ===========================================
-- DATA SUMMARY
-- ===========================================

-- Display summary of loaded data
SELECT
    (SELECT COUNT(*) FROM users) as total_users,
    (SELECT COUNT(*) FROM users WHERE role = 'ADMIN') as admin_users,
    (SELECT COUNT(*) FROM users WHERE role = 'USER') as regular_users,
    (SELECT COUNT(*) FROM events) as total_events,
    (SELECT COUNT(*) FROM events WHERE starts_at > NOW()) as upcoming_events,
    (SELECT COUNT(*) FROM events WHERE starts_at <= NOW()) as past_events,
    (SELECT COUNT(*) FROM bookings) as total_bookings,
    (SELECT COUNT(*) FROM bookings WHERE status = 'CONFIRMED') as confirmed_bookings,
    (SELECT COUNT(*) FROM bookings WHERE status = 'PENDING') as pending_bookings,
    (SELECT COUNT(*) FROM bookings WHERE status = 'CANCELLED') as cancelled_bookings;
