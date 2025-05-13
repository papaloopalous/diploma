local function init_storage()
-- Create session space
    box.schema.space.create('sessions', {
        format = {
            {name = 'session_id', type = 'string'},
            {name = 'user_id', type = 'string'},
            {name = 'role', type = 'string'}
        },
        if_not_exists = true
    })

    -- Create primary index on session_id
    box.space.sessions:create_index('primary', {
        parts = {1, 'string'},
        if_not_exists = true
    })

    -- Grant permissions
    box.schema.user.grant('guest', 'read,write,execute', 'space', 'sessions')
    end
end

return init_storage