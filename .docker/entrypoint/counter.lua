box.cfg {
    listen = 3302;
    memtx_memory = 1024 * 1024 * 1024 * 2;
    net_msg_max = 100000,
    readahead = 100000
}


box.once('init', function()
     -- box.schema.user.grant('guest', 'read,write,execute', 'universe')
end)
box.once('schema', function()
    if not box.space.message_unread_counter then
        local unread_counter = box.schema.space.create('message_unread_counter')
        unread_counter:format({
            {name = 'id', type = 'unsigned'},
            {name = 'from', type = 'unsigned'},
            {name = 'to', type = 'unsigned'},
            {name = 'count', type = 'unsigned'},
        })

         unread_counter:create_index('primary', {
            parts = {'id'},
            type = 'tree',
            unique = true,
            sequence = true
         })

        unread_counter:create_index('from_to', {
            parts = {'from', 'to'},
            unique = false
        })
    end
end)

function create_counter(from, to)
    local unread_counter = box.space.message_unread_counter:auto_increment{from, to, 1}
    return unread_counter
end

function increment_counter(from, to)
    -- Select returns a list of tuples, even if it's empty
    local unread_counters = box.space.message_unread_counter.index.from_to:select({from, to})

    if #unread_counters == 0 then
        -- If no counter exists, create a new one
        create_counter(from, to)
    else
        -- Assuming the first counter is the one we want (if duplicates are not expected)
        local unread_counter = unread_counters[1]
        -- Increment the count. Note: '+', 4, 1 performs increment without needing to read the current value
        box.space.message_unread_counter:update(unread_counter.id, {{'+', 4, 1}})
    end
end

function get_unread_counter(from,to)
    local unread_counter = box.space.message_unread_counter.index.from_to:select({from, to})
    return unread_counter
end

function decrement_counter(from, to, count)
    local unread_counters = box.space.message_unread_counter.index.from_to:select({from, to})
    if #unread_counters > 0 then
        local unread_counter = unread_counters[1]
        -- Ensure the decrement does not result in a negative number
        local new_count = math.max(0, unread_counter[4] - count)
        -- Update the counter safely
        box.space.message_unread_counter:update(unread_counter.id, {{'=', 4, new_count}})
    end
end

box.schema.func.create('create_counter', {if_not_exists = true})
box.schema.func.create('increment_counter', {if_not_exists = true})
box.schema.func.create('get_unread_counter', {if_not_exists = true})
box.schema.func.create('decrement_counter', {if_not_exists = true})
box.schema.func.create('reset_counter', {if_not_exists = true})


box.schema.user.grant('guest', 'execute', 'function', 'create_counter', {if_not_exists = true})
box.schema.user.grant('guest', 'execute', 'function', 'increment_counter', {if_not_exists = true})
box.schema.user.grant('guest', 'execute', 'function', 'get_unread_counter', {if_not_exists = true})
box.schema.user.grant('guest', 'execute', 'function', 'decrement_counter', {if_not_exists = true})
box.schema.user.grant('guest', 'execute', 'function', 'reset_counter', {if_not_exists = true})

require('console').start()
