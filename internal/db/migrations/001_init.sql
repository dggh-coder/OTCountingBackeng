CREATE TABLE IF NOT EXISTS work_session (
    session_id BIGINT PRIMARY KEY,
    work_date DATE NOT NULL,
    period VARCHAR(2) NOT NULL CHECK (period IN ('AM','PM')),
    status VARCHAR(16) NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (work_date, period)
);

CREATE TABLE IF NOT EXISTS time_entry (
    entry_id VARCHAR(64) NOT NULL,
    session_id BIGINT NOT NULL REFERENCES work_session(session_id) ON DELETE CASCADE,
    employee_id VARCHAR(8) NOT NULL CHECK (employee_id IN ('A','B')),
    entry_type VARCHAR(8) NOT NULL CHECK (entry_type IN ('OT','BREAK')),
    start_time VARCHAR(5) NOT NULL,
    end_time VARCHAR(5) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (session_id, entry_id)
);

CREATE TABLE IF NOT EXISTS session_result (
    session_id BIGINT NOT NULL REFERENCES work_session(session_id) ON DELETE CASCADE,
    employee_id VARCHAR(8) NOT NULL CHECK (employee_id IN ('A','B')),
    date_label VARCHAR(16) NOT NULL,
    rate20_minutes INTEGER NOT NULL,
    rate20_rounded_hours INTEGER NOT NULL,
    rate15_minutes INTEGER NOT NULL,
    rate15_rounded_hours INTEGER NOT NULL,
    total_ot_minutes INTEGER NOT NULL,
    total_break_minutes INTEGER NOT NULL,
    net_work_minutes INTEGER NOT NULL,
    calculated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (session_id, employee_id)
);

CREATE TABLE IF NOT EXISTS session_rendered_fragment (
    session_id BIGINT NOT NULL REFERENCES work_session(session_id) ON DELETE CASCADE,
    employee_id VARCHAR(8) NOT NULL,
    fragment_type VARCHAR(64) NOT NULL,
    format_version INTEGER NOT NULL,
    content_html TEXT NOT NULL,
    last_calculated TIMESTAMP NOT NULL,
    PRIMARY KEY (session_id, employee_id, fragment_type, format_version)
);
