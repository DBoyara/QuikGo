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
local logfile = io.open(script_path .. "\\logs\\script_log.txt", "a")
function log(msg, level)
    level = level or 1
    local log_msg = os.date("%Y-%m-%d %H:%M:%S") .. " [LOG " .. level .. "]: " .. msg

    -- Запись в файл
    if logfile then
        logfile:write(log_msg .. "\n")
        logfile:flush()
    end

    -- Вывод в окно "Лог Lua"
    PrintDbgStr(log_msg)
end

-- Создаём сокет-сервер
function setup_socket()
    local attempts = 3
    local delay_between_attempts = 1000  -- Задержка между попытками в миллисекундах

    for i = 1, attempts do
        local status, server = pcall(socket.bind, response_host, response_port)
        if status and server then
            server:settimeout(0)
            log("Сокет-сервер запущен на " .. response_host .. ":" .. response_port, 1)
            return server
        else
            log("Ошибка при создании сокет-сервера (попытка " .. i .. " из " .. attempts .. "): " .. (server or "неизвестная ошибка"), 3)
            if i < attempts then
                sleep(delay_between_attempts)
            end
        end
    end

    log("Не удалось создать сокет-сервер после " .. attempts .. " попыток", 3)
    return nil
end

-- Принимаем подключение клиента
function accept_client()
    local status, client = pcall(response_server.accept, response_server)
    if status and client then
        is_connected = true
        client:settimeout(0)
        log("Клиент подключён", 1)
        return client
    elseif not status then
        -- Это реальная ошибка (например, сокет закрыт или недоступен)
        log("Ошибка при подключении клиента: " .. (client or "неизвестная ошибка"), 3)
    else
        -- Это штатная ситуация: клиент не подключился, но сокет работает нормально
        -- Не логируем это как ошибку
    end
    return nil
end

-- Получаем сообщения
function receive_message(client)
    local msg, err = client:receive("*l")
    if msg then
        log("Получено сообщение: " .. msg, 0)
        local status, decoded_msg = pcall(json.decode, msg)
        if status then
            return decoded_msg
        else
            log("Ошибка декодирования JSON: " .. decoded_msg, 3)
            return nil, "Ошибка декодирования JSON"
        end
    elseif err ~= "timeout" then
        -- Это реальная ошибка (например, клиент отключился)
        log("Ошибка получения сообщения: " .. err, 3)
        return nil, err
    else
        -- Это штатная ситуация: клиент не отправил данные (таймаут)
        -- Не логируем это как ошибку
        return nil
    end
end

-- Отправляем ответ
function send_response(client, response)
    local response_str = json.encode(response)
    if response_str then
        local attempts = 3
        for i = 1, attempts do
            local status, err = pcall(client.send, client, response_str .. "\n")
            if status then
                log("Отправлен ответ: " .. response_str, 0)
                return true
            else
                log("Ошибка отправки ответа (попытка " .. i .. " из " .. attempts .. "): " .. (err or "неизвестная ошибка"), 3)
                if i < attempts then
                    sleep(1000)  -- Задержка между попытками
                end
            end
        end
        log("Не удалось отправить ответ после " .. attempts .. " попыток", 3)
        return false
    else
        log("Ошибка кодирования JSON для ответа", 3)
        return false
    end
end

-- Создаём DataSource
function create_data_source(class_code, sec_code, interval)
    local attempts = 3
    local delay_between_attempts = 1000  -- Задержка между попытками в миллисекундах

    for i = 1, attempts do
        local status, ds, error_msg = pcall(CreateDataSource, class_code, sec_code, interval)
        if status and ds then
            return ds
        else
            log("Ошибка создания DataSource для " .. sec_code .. " (попытка " .. i .. " из " .. attempts .. "): " .. (error_msg or "неизвестная ошибка"), 3)
            if i < attempts then
                sleep(delay_between_attempts)
            end
        end
    end

    return nil, "Не удалось создать DataSource после " .. attempts .. " попыток"
end

-- Обрабатываем команду создания DataSource
function handle_create_data_source(data)
    if not data.ticker or not data.interval then
        return { success = false, message = "Не указаны ticker или interval" }
    end

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
    if not data.ticker or not data.interval then
        return { success = false, message = "Не указаны ticker или interval" }
    end

    local sec_code = data.ticker
    local interval = data.interval
    local count = data.count or 10
    count = math.min(count, 1000)  -- Ограничение на 1000 свечей

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
function process_request(request)
    if request.cmd == "ping" then
        log("Получена команда PING", 0)
        return { cmd = "pong", message = "Pong from QUIK" }
    elseif request.cmd == "create_data_source" then
        return handle_create_data_source(request.data)
    elseif request.cmd == "get_candles" then
        return handle_get_candles(request.data)
    else
        log("Неизвестная команда: " .. tostring(request.cmd), 2)
        return { success = false, message = "Неизвестная команда" }
    end
end

function main()
    response_server = setup_socket()
    if not response_server then
        log("Ошибка при настройке сокета", 3)
        return
    end

    while true do
        if not is_connected then
            response_client = accept_client()
        else
            local request, err = receive_message(response_client)
            if request then
                local response = process_request(request)
                send_response(response_client, response)
            elseif err then
                if err ~= "timeout" then
                    -- Это реальная ошибка (например, клиент отключился)
                    log("Клиент отключился: " .. err, 3)
                    response_client = nil
                    is_connected = false
                else
                    -- Это штатная ситуация: клиент не отправил данные (таймаут)
                    -- Не логируем это как ошибку
                end
            end
        end
        sleep(100)
    end
end