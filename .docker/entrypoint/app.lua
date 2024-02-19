box.cfg {
    listen = 3301;
    memtx_memory = 1024 * 1024 * 1024 * 2; -- 1 GB
    net_msg_max = 100000, -- Adjust this value as necessary
    readahead = 100000 -- Adjust this value as necessary
}

box.once('init', function()
    -- box.schema.user.grant('guest', 'read,write,execute', 'universe')
end)

if not box.space.dialogues then
    local dialogues = box.schema.space.create('dialogues')
    dialogues:format({
        {name = 'id', type = 'unsigned'},
        {name = 'from', type = 'unsigned'},
        {name = 'to', type = 'unsigned'},
        {name = 'created_at', type = 'unsigned'}
    })

     dialogues:create_index('primary', {
        parts = {'id'},
        type = 'tree',
        unique = true,
        sequence = true
     })


    dialogues:create_index('from_to', {
        parts = {'from', 'to'},
        unique = false
    })
end

if not box.space.messages then
    local messages = box.schema.space.create('messages')
    messages:format({
        {name = 'id', type = 'unsigned'},
        {name = 'dialogue_id', type = 'unsigned'},
        {name = 'from', type = 'unsigned'},
        {name = 'to', type = 'unsigned'},
        {name = 'message', type = 'string'},
        {name = 'created_at', type = 'unsigned'}
    })

    messages:create_index('primary', {
        parts = {'id'},
        type = 'tree',
        unique = true,
        sequence = true
    })

    messages:create_index('dialogue_messages_index', {
            parts = {'dialogue_id'},
            unique = false
        })
end

-- box.schema.space.create('test')
-- box.space.test:create_index('primary')

function create_dialogue(from, to)
    created_at = os.time()
    local dialogue_id = box.space.dialogues:auto_increment{from, to, created_at}
    return dialogue_id
end

function create_message(dialogue_id, from, to, message)
    created_at = os.time()

    box.space.messages:auto_increment{dialogue_id, from, to, message, created_at}
end

function is_dialogue_exist(from, to)
    local dialogue = box.space.dialogues.index['from_to']:select({from, to})[1]
    if dialogue then
        return dialogue[1]
    else
        return nil
    end
end

function get_dialogue(user_id, with_user_id)
    -- Call the is_dialogue_exist function to get the dialogue ID if it exists
    local dialogue_id = is_dialogue_exist(user_id, with_user_id)

    -- Check if a dialogue ID exists by looking for a non-nil value
    if dialogue_id then
        -- Use the dialogue_id to select messages by dialogue_id
        local messages = box.space.messages.index['dialogue_messages_index']:select(dialogue_id)
        return messages
    else
        return nil
    end
end


box.schema.func.create('create_dialogue', {if_not_exists = true})
box.schema.func.create('create_message', {if_not_exists = true})
box.schema.func.create('is_dialogue_exist', {if_not_exists = true})
box.schema.func.create('get_dialogue', {if_not_exists = true})

box.schema.user.grant('guest', 'execute', 'function', 'get_dialogue', {if_not_exists = true})
box.schema.user.grant('guest', 'execute', 'function', 'create_dialogue', {if_not_exists = true})
box.schema.user.grant('guest', 'execute', 'function', 'is_dialogue_exist', {if_not_exists = true})
box.schema.user.grant('guest', 'execute', 'function', 'create_message', {if_not_exists = true})


require('console').start()