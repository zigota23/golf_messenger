CREATE TABLE ttrs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    course_name VARCHAR(255) NOT NULL,
    course_location VARCHAR(255),
    tee_date DATE NOT NULL,
    tee_time TIME NOT NULL,
    max_players INTEGER DEFAULT 4,
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    captain_user_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(50) DEFAULT 'OPEN',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_ttrs_tee_date ON ttrs(tee_date);
CREATE INDEX idx_ttrs_captain ON ttrs(captain_user_id);

CREATE TABLE ttr_co_captains (
    ttr_id UUID REFERENCES ttrs(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (ttr_id, user_id)
);

CREATE TABLE ttr_players (
    ttr_id UUID REFERENCES ttrs(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'CONFIRMED',
    PRIMARY KEY (ttr_id, user_id)
);

CREATE TABLE invitations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ttr_id UUID NOT NULL REFERENCES ttrs(id) ON DELETE CASCADE,
    inviter_user_id UUID NOT NULL REFERENCES users(id),
    invitee_user_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(50) DEFAULT 'PENDING',
    message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP NULL,
    UNIQUE(ttr_id, invitee_user_id)
);

CREATE INDEX idx_invitations_invitee ON invitations(invitee_user_id);
CREATE INDEX idx_invitations_status ON invitations(status);
