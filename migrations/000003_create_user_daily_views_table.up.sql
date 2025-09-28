CREATE TABLE user_daily_views (
    viewer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shown_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    view_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    
    PRIMARY KEY (viewer_id, shown_user_id, view_date)
);

CREATE INDEX idx_user_daily_views_viewer_date ON user_daily_views(viewer_id, view_date);
CREATE INDEX idx_user_daily_views_date ON user_daily_views(view_date);
CREATE INDEX idx_user_daily_views_shown_user ON user_daily_views(shown_user_id);