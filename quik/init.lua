-- Устанавливаем пути для библиотек
script_path = getScriptPath()
package.path = package.path .. ";" .. script_path .. "\\?.lua;" .. script_path .. "\\?.luac"..";"..".\\?.lua;"..".\\?.luac"
package.cpath = package.cpath .. ";" .. script_path .. "\\?.dll" .. ";" .. ".\\?.dll"

local socket = require("socket")
local json = require("dkjson")

local response_host = "127.0.0.1"
local response_port = 34130
local response_server
local response_client
local is_connected = false

-- Хранилище DataSource
local data_sources = {}

-- Логирование
function log(msg, level)
    level = level or 1
    local log_msg = os.date("%Y-%m-%d %H:%M:%S") .. " [LOG " .. level .. "]: " .. msg
    print(log_msg)
end

-- Создаём сокет-сервер
function setup_socket()
    response_server = socket.bind(response_host, response_port)
    if not response_server then
        log("Не удалось создать сокет-сервер на " .. response_host .. ":" .. response_port, 3)
        return false
    end
    response_server:settimeout(0)
    log("Сокет-сервер запущен на " .. response_host .. ":" .. response_port, 1)
    return true
end

-- Принимаем подключение клиента
function accept_client()
    local client = response_server:accept()
    if client then
        is_connected = true
        client:settimeout(0)
        log("Клиент подключён", 1)
        return client
    end
    return nil
end

-- Получаем сообщения
function receive_message(client)
    local msg, err = client:receive("*l")
    if msg then
        log("Получено сообщение: " .. msg, 0)
        local decoded_msg = json.decode(msg)
        return decoded_msg
    elseif err ~= "timeout" then
        log("Ошибка получения сообщения: " .. err, 3)
        return nil, err
    end
end

-- Отправляем ответ
function send_response(client, response)
    local response_str = json.encode(response)
    if response_str then
        client:send(response_str .. "\n")
        log("Отправлен ответ: " .. response_str, 0)
    end
end

-- Создаём DataSource
function create_data_source(class_code, sec_code, interval)
    local ds, error_msg = CreateDataSource(class_code, sec_code, interval)
    if not ds then
        return nil, "Ошибка создания DataSource для " .. sec_code .. ": " .. error_msg
    end
    return ds
end

-- Обрабатываем команду создания DataSource
function handle_create_data_source(data)
    local class_code = data.class_code or "TQBR"
    local sec_code = data.ticker
    local interval = data.interval

    local key = sec_code .. "|" .. interval
    if data_sources[key] then
        return { success = true, message = "DataSource уже существует." }
    end

    local ds, error_msg = create_data_source(class_code, sec_code, interval)
    if ds then
        data_sources[key] = ds
        return { success = true, message = "DataSource создан." }
    else
        return { success = false, message = error_msg }
    end
end

-- Получение свечей
function handle_get_candles(data)
    local sec_code = data.ticker
    local interval = data.interval
    local count = data.count or 10

    local key = sec_code .. "|" .. interval
    local ds = data_sources[key]
    if not ds then
        return { success = false, message = "DataSource для " .. sec_code .. " не существует." }
    end

    local size = ds:Size()
    local candles = {}
    for i = math.max(1, size - count + 1), size do
        table.insert(candles, {
            time = ds:T(i),
            open = ds:O(i),
            high = ds:H(i),
            low = ds:L(i),
            close = ds:C(i),
            volume = ds:V(i)
        })
    end
    return { success = true, candles = candles }
end

-- Главный цикл
function main()
    if not setup_socket() then
        log("Ошибка при настройке сокета", 3)
        return
    end

    while true do
        if not is_connected then
            response_client = accept_client()
        else
            local request, err = receive_message(response_client)
            if request then
                if request.cmd == "ping" then
                    log("Получена команда PING", 0)
                    local response = { cmd = "pong", message = "Pong from QUIK" }
                elseif request.cmd == "create_data_source" then
                    local response = handle_create_data_source(request.data)
                    send_response(response_client, response)
                elseif request.cmd == "get_candles" then
                    local response = handle_get_candles(request.data)
                    send_response(response_client, response)
                else
                    log("Неизвестная команда: " .. tostring(request.cmd), 2)
                end
            elseif err then
                log("Клиент отключился", 3)
                response_client = nil
                is_connected = false
            end
        end
        sleep(100)
    end
end
