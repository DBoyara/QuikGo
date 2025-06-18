--~ Copyright (c) 2025 QuikGO Author https://github.com/DBoyara. All rights reserved.
--~ Licensed under the Apache License, Version 2.0. See LICENSE.txt in the project root for license information.

-- Устанавливаем пути для библиотек
script_path = getScriptPath()
package.path = package.path .. ";" .. script_path .. "\\?.lua;" .. script_path .. "\\?.luac"..";"..".\\?.lua;"..".\\?.luac"
package.cpath = package.cpath .. ";" .. script_path .. "\\?.dll" .. ";" .. ".\\?.dll"

local socket = require("socket")
local json = require("dkjson")

local response_host = "127.0.0.1"
local response_port = 54320
local callback_host = "127.0.0.1"
local callback_port = 54321
local response_server
local response_client
local is_connected = false
local is_subscribed = false
local INTERVAL_TICKS = 1

-- Хранилище DataSource
local data_sources = {}

-- Логирование
local logfile = io.open(script_path .. "\\logs\\script_log.txt", "a")
local request_id = 0
function log(msg, level)
    level = level or 1
    request_id = request_id + 1
    local log_msg = os.date("%Y-%m-%d %H:%M:%S") .. " [LOG " .. level .. "] [ReqID: " .. request_id .. "]: " .. msg

    if logfile then
        logfile:write(log_msg .. "\n")
        logfile:flush()
    end

    PrintDbgStr(log_msg)
end

-- Подключаемся к Go-серверу
function connect_to_go()
    local client, err = socket.tcp()
    if not client then
        log("Ошибка создания сокета: " .. err, 3)
        return nil
    end

    local success, err = client:connect(callback_host, callback_port)
    if not success then
        log("Ошибка подключения к Go: " .. err, 3)
        return nil
    end

    log("Подключение к Go-серверу установлено!", 1)

    client:settimeout(0)
    return client
end

local event_client = connect_to_go()


function to_json(msg)
    local status, str = pcall(json.encode, msg, { indent = false })

    if status then
        return str
    else
        log("Ошибка сериализации JSON: " .. tostring(str), 3)
        return nil
    end
end

-- Подписка на события QUIK
function subscribe_to_callbacks()
    is_subscribed = true
    log("Подписка на события QUIK активирована", 1)
end

-- Создаём сокет-сервер
function setup_socket()
    local attempts = 3
    local delay_between_attempts = 1000

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
    local status, response_str = pcall(json.encode, response)
    if not status then
        log("Ошибка кодирования JSON: " .. response_str, 3)
        return false
    end

    local attempts = 3
    for i = 1, attempts do
        local status, err = pcall(client.send, client, response_str .. "\n")
        if status then
            log("Отправлен ответ: " .. response_str, 0)
            return true
        else
            log("Ошибка отправки ответа (попытка " .. i .. " из " .. attempts .. "): " .. (err or "неизвестная ошибка"), 3)
            if i < attempts then
                sleep(1000)
            end
        end
    end
    log("Не удалось отправить ответ после " .. attempts .. " попыток", 3)
    return false
end

-- Функция для получения последней цены тикера
function get_last_price(class_code, sec_code)
    local param = getParamEx(class_code, sec_code, "LAST")
    if param and param.param_value then
        return { success = true, message = param.param_value }
    else
        return { success = false, message = "Не удалось получить последнюю цену для " .. sec_code }
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

-- Обрабатываем команду закрытия DataSource
function handle_close_data_source(data)
    if not data.ticker or not data.interval then
        return { success = false, message = "Не указаны ticker или interval" }
    end

    local sec_code = data.ticker
    local interval = data.interval
    local key = sec_code .. "|" .. interval

    if not data_sources[key] then
        return { success = false, message = "DataSource для " .. key .. " не найден." }
    end

    -- Закрываем DataSource
    data_sources[key]:Close()
    data_sources[key] = nil  -- Удаляем из таблицы

    return { success = true, message = "DataSource для " .. key .. " успешно закрыт." }
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

-- Получение списка всех торговых счетов
function get_trade_accounts()
    local trade_accounts = {}
    local count = getNumberOf("trade_accounts")
    for i = 0, count - 1 do
        local account = getItem("trade_accounts", i)
        table.insert(trade_accounts, account)
    end
    return trade_accounts
end

-- Получение торгового счета по коду класса
function get_trade_account(class_code)
    local count = getNumberOf("trade_accounts")
    for i = 0, count - 1 do
        local account = getItem("trade_accounts", i)
        if string.find(account.class_codes, class_code) then
            return account
        end
    end
    return nil
end

-- Функция возвращает информацию по всем денежным лимитам
function get_money_limits()
    local money_limits = {}
    local count = getNumberOf("money_limits")
    for i = 0, count - 1 do
        local limit = getItem("money_limits", i)
        table.insert(money_limits, limit)
    end
    return money_limits
end

-- Функция возвращает информацию о клиентском портфеле
function get_portfolio_info(firmid, client_code)
    local portfolio = getPortfolioInfo(firmid, client_code)
    return portfolio
end

-- Отправка торговой транзакции
function send_transaction(transaction_params)
    local result = sendTransaction(transaction_params)
    if result == "" then
        return { success = true, message = "transaction send success" }
    else
        return { success = false, message = "error send transaction: " .. result }
    end
end

-- Функция возвращает заявку по номеру и классу инструмента
function get_order_by_number(class_code, order_id)
    local order = getOrderByNumber(class_code, order_id)
    return order
end

-- Функция возвращает заявку по заданному инструменту и ID-транзакции
function get_order_by_id(class_code, sec_code, trans_id)
    local order_num = 0
    local selected_order = nil

    for i = 0, getNumberOf("orders") - 1 do
        local order = getItem("orders", i)
        if order.class_code == class_code and order.sec_code == sec_code and order.trans_id == tonumber(trans_id) then
            if order.order_num > order_num then
                order_num = order.order_num
                selected_order = order
            end
        end
    end

    return selected_order
end

-- Функция возвращает список стоп-заявок по заданному инструменту
function get_stop_orders(class_code, sec_code)
    local stop_orders = {}

    for i = 0, getNumberOf("stop_orders") - 1 do
        local stop_order = getItem("stop_orders", i)
        if stop_order.class_code == class_code and stop_order.sec_code == sec_code then
            table.insert(stop_orders, stop_order)
        end
    end

    return stop_orders
end

-- Callbacks
function OnOrder(order)
    send_event("OnOrder", order)
end

function OnTrade(trade)
    send_event("OnTrade", trade)
end

function OnAllTrade(alltrade)
    send_event("OnAllTrade", alltrade)
end

function OnTransReply(trans_reply)
    send_event("OnTransReply", trans_reply)
end

function OnQuote(class_code, sec_code)
    local msg = {}
    local status, ql2 = pcall(getQuoteLevel2, class_code, sec_code)
    if status then
        msg = ql2
        msg.class_code = class_code
        msg.sec_code = sec_code

        send_event("OnQuote", msg)
    else
        log("Ошибка при получении стакана для " .. sec_code, 3)
    end
end

function OnConnected()
    send_event("OnConnected", "QUIK подключен")
end

function OnDisconnected()
    send_event("OnDisconnected", "QUIK отключен")
end

function OnStopOrder(stop_order)
    send_event("OnStopOrder", stop_order)
end

function OnAccountBalance(acc_bal)
    send_event("OnAccountBalance", acc_bal)
end

function OnAccountPosition(acc_pos)
    send_event("OnAccountPosition", acc_pos)
end

function OnFirm(firm)
    send_event("OnFirm", firm)
end

function OnFuturesLimitChange(fut_limit)
    send_event("OnFuturesLimitChange", fut_limit)
end

function OnFuturesLimitDelete(lim_del)
    send_event("OnFuturesLimitDelete", lim_del)
end

function OnFuturesClientHolding(fut_pos)
    send_event("OnFuturesClientHolding", fut_pos)
end

function OnMoneyLimit(mlimit)
    send_event("OnMoneyLimit", mlimit)
end

function OnMoneyLimitDelete(mlimit_del)
    send_event("OnMoneyLimitDelete", mlimit_del)
end

function OnDepoLimit(dlimit)
    send_event("OnDepoLimit", dlimit)
end

function OnDepoLimitDelete(dlimit_del)
    send_event("OnDepoLimitDelete", dlimit_del)
end

function OnStop()
    send_event("OnStop", "Lua завершает работу")
end

function OnClose()
    send_event("OnClose", "QUIK# завершает работу")
end

function OnInit(script_path)
    send_event("OnInit", script_path)
    log("QUIK# is initialized from "..script_path, 0)
end

-- Функция для подписки на обновления стакана котировок
function subscribe_to_order_book(class_code, sec_code)
    local key = sec_code .. "|ORDERBOOK"
    if data_sources[key] then
        log("Подписка на стакан для " .. sec_code .. " уже существует.", 2)
        return
    end

    -- Создаем DataSource для стакана котировок
    local ds, err = create_data_source(class_code, sec_code, INTERVAL_TICKS)
    if not ds then
        log("Ошибка при создании DataSource для стакана " .. sec_code .. ": " .. err, 3)
        return
    end

    -- Сохраняем DataSource в таблицу
    data_sources[key] = ds

    -- Callback-функция для обработки обновлений стакана
    function OnData(ds)
        local msg = {}
        local status, ql2 = pcall(getQuoteLevel2, class_code, sec_code)
        if status then
            msg = ql2
            msg.class_code = class_code
            msg.sec_code = sec_code

            -- Отправляем данные клиенту
            send_event("OrderBookUpdate", msg)
        else
            log("Ошибка при получении стакана для " .. sec_code, 3)
        end
    end

    -- Подписываемся на обновления
    ds:SetUpdateCallback(OnData)

    log("Подписка на стакан для " .. sec_code .. " успешно создана.", 1)
end

-- Функция отправки событий
function send_event(event_name, event_data)
    if not event_client then
        log("Соединение с Go отсутствует!", 3)
        return
    end

    if event_name == "OnQuote" then
        local msg = json.encode({ cmd = event_name, order_book = event_data }) .. "\n"
    elseif event_name == "OrderBookUpdate " then
        local msg = json.encode({ cmd = event_name, order_book = event_data }) .. "\n"
    else
        local msg = json.encode({ cmd = event_name, data = event_data }) .. "\n"
    end

    local success, err = event_client:send(msg)

    if success then
        log("Отправлено событие в Go: " .. event_name, 1)
    else
        log("Ошибка отправки события: " .. err, 3)
    end
end

-- Главный цикл
function process_request(request)
    if request.cmd == "ping" then
        log("Get PING", 0)
        return { cmd = "pong", message = "Pong from QUIK" }
    elseif request.cmd == "createDataSource" then
        return handle_create_data_source(request.data)
    elseif request.cmd == "closeDataSource" then
        return handle_close_data_source(request.data)
    elseif request.cmd == "subscribeOrderBook" then
        return subscribe_to_order_book(request.data.class_code, request.data.sec_code)
    elseif request.cmd == "getСandles" then
        return handle_get_candles(request.data)
    elseif request.cmd == "getTradeAccounts" then
        local accounts = get_trade_accounts()
        return { success = true, accounts = accounts }
    elseif request.cmd == "getTradeAccount" then
        if not request.data or not request.data.class_code then
            return { success = false, message = "no class_code" }
        end
        local account = get_trade_account(request.data.class_code)
        if account then
            return { success = true, account = account }
        else
            return { success = false, message = "trade account not found" }
        end
    elseif request.cmd == "getMoneyLimits" then
        local limits = get_money_limits()
        return { success = true, limits = limits }
    elseif request.cmd == "getPortfolioInfo" then
        if not request.data or not request.data.firmId or not request.data.clientCode then
            return { success = false, message = "no firmId or clientCode" }
        end
        local portfolio_info = get_portfolio_info(request.data.firmId, request.data.clientCode)
        if portfolio_info then
            return { success = true, portfolio = portfolio_info }
        else
            return { success = false, message = "can't get info about portfolio" }
        end
    elseif request.cmd == "sendTransaction" then
        if not request.data  then
            return { success = false, message = "no transaction data" }
        end
        local result = send_transaction(request.data)
        return result
    elseif request.cmd == "getOrderByNumber" then
        if not request.data or not request.data.class_code or not request.data.order_id then
            return { success = false, message = "Не указаны class_code или order_id" }
        end
        local order = get_order_by_number(request.data.class_code, request.data.order_id)
        if order then
            return { success = true, order = order }
        else
            return { success = false, message = "Заявка не найдена" }
        end
    elseif request.cmd == "getOrderById" then
        if not request.data or not request.data.class_code or not request.data.sec_code or not request.data.trans_id then
            return { success = false, message = "Не указаны class_code или sec_code или trans_id" }
        end
        local order = get_order_by_id(request.data.class_code, request.data.sec_code, request.data.trans_id)
        if order then
            return { success = true, order = order }
        else
            return { success = false, message = "Заявка не найдена" }
        end
    elseif request.cmd == "getStopOrders" then
        if not request.data or not request.data.class_code or not request.data.sec_code then
            return { success = false, message = "Не указаны class_code или sec_code" }
        end
        local order = get_stop_orders(request.data.class_code, request.data.sec_code)
        if order then
            return { success = true, order = order }
        else
            return { success = false, message = "Заявка не найдена" }
        end
    elseif request.cmd == "subscribeToEvents" then
        subscribe_to_callbacks()
        return { success = true, message = "Подписка на события активирована" }
    elseif request.cmd == "getLastPrice" then
        if not request.data or not request.data.class_code or not request.data.sec_code then
            return { success = false, message = "Не указаны class_code или sec_code" }
        end
        return get_last_price(request.data.class_code, request.data.sec_code)
    else
        log("not no this command: " .. tostring(request.cmd), 2)
        return { success = false, message = "not no this command" }
    end
end

-- Функция для переподключения к Go-серверу
function reconnect_to_go(max_attempts, delay)
    max_attempts = max_attempts or 5  -- Максимальное количество попыток (по умолчанию 5)
    delay = delay or 2000             -- Задержка между попытками (в миллисекундах)

    for attempt = 1, max_attempts do
        log("Попытка подключения #" .. attempt, 2)

        event_client = connect_to_go()

        if event_client then
            log("Успешное подключение к Go-серверу!", 1)
            return true
        end

        log("Не удалось подключиться. Повтор через " .. (delay / 1000) .. " секунд.", 3)
        sleep(delay) -- Ожидание перед следующей попыткой
    end

    log("Ошибка: не удалось подключиться после " .. max_attempts .. " попыток.", 3)
    return false
end

-- Основная функция запуска сервера и обработки запросов.
function main()
    response_server = setup_socket()

    if not response_server then
        log("Ошибка при настройке сокета", 3)
        return
    end

    while true do
        if not is_connected then
            response_client = accept_client()

            if not response_client then
                sleep(500) -- Ждём перед повторной проверкой подключения клиента.
            else
                is_connected = true
                log("Клиент успешно подключен",1)

                -- Проверяем соединение с Go-сервером и переподключаемся при необходимости.
                if not event_client or not reconnect_to_go() then
                    log("Не удалось установить соединение с Go. Завершение работы.",3)
                    break
                end
            end

        else
            local request, err = receive_message(response_client)

            if request then
                local response = process_request(request)
                send_response(response_client, response)

            elseif err and err ~= "timeout" then
               -- Ошибка связи или отключение клиента.
               log ("Клиент отключился: "..err ,3 )
               response_client=nil
               is_connected=false
           end
       end

       sleep(100)   -- Пауза перед следующим циклом обработки событий.
   end
end
