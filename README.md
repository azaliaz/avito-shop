# **Магазин мерча**

Сервис, который позволяет сотрудникам обмениваться монетками и использовать их для покупки мерча. В рамках этого сервиса сотрудники могут:
- Просматривать список купленных ими мерчевых товаров.
- Получать детализированную информацию о перемещении монеток в их кошельке.
- Также в сервисе предусмотрены ограничения:
    - Баланс монеток не может быть отрицательным, что исключает возможность ухода в минус при операциях.
    - Все транзакции проверяются на достаточность средств перед выполнением.
Система обеспечивает удобный и прозрачный учет монеток, помогая сотрудникам эффективно управлять своими внутренними активами.

**Мерч** — это продукт, который можно купить за монетки. Всего в магазине доступно 10 видов мерча. Каждый товар имеет уникальное название и цену. Ниже приведён список наименований и их цены.

| Название     | Цена |
|--------------|------|
| t-shirt      | 80   |
| cup          | 20   |
| book         | 50   |
| pen          | 10   |
| powerbank    | 200  |
| hoody        | 300  |
| umbrella     | 200  |
| socks        | 10   |
| wallet       | 50   |
| pink-hoody   | 500  |

Предполагается, что в магазине бесконечный запас каждого вида мерча.

## Используемые технологии:

- PostgreSQL (в качестве хранилища данных)
- Docker (для запуска сервиса)
- Fiber (веб фреймворк)
- golang-migrate/migrate (для миграций БД)
- pgx (драйвер для работы с PostgreSQL)
- golang/mock, testify (для тестирования)

## Запуск

Запустить сервис можно с помощью команды `make deploy`

Останавливает и удаляет развернутое окружение Docker вместе с его образами `make rollback`
Для запуска unit-тестов необходимо выполнить команду `make unit_test`

Для запуска тестов с покрытием `make cover` и `make cover-html` для получения отчёта в html формате

Для запуска интеграционных тестов необходимо выполнить команду `make integration_tests`

Для запуска линтера необходимо выполнить команду `make lint`


## Примеры запросов

- [Аутентификация](#sign-in)
- [Купить предмет за монеты](#buy-item)
- [Отправить монеты другому пользователю](#send-coin)
- [Получить информацию о монетах, инвентаре и истории транзакций](#get-info)


### Аутентификация <a name="sign-in"></a>
```curl -X POST http://localhost:8080/api/auth \
     -H "Content-Type: application/json" \
     -d '{
           "username": "user_1",
           "password": "password123"
         }'
```

Пример ответа:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.Cu0jgoMXBW7FiljBNqT63i0TKnwWUFBfTcTDJiMXWO
```
Вывод в таблице `users`:

![Запись в базе данных пользователей после аутентификации](image/image_1.png)

### Купить предмет за монеты <a name="buy-item"></a>

``` curl -X GET http://localhost:8080/api/buy/book \
     -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.Cu0jgoMXBW7FiljBNqT63i0TKnwWUFBfTcTDJiMXWOg" \
     -H "Content-Type: application/json"
```

Пример ответа:

``` 200 OK ```

Вывод в таблице `inventory` после покупки мерча (book) user_1:

![Вывод таблицы c информацией о покупках](image/image_2.png)

### Отправить монеты другому пользователю <a name="send-coin"></a>

``` curl -X POST http://localhost:8080/api/sendCoin \
     -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.Cu0jgoMXBW7FiljBNqT63i0TKnwWUFBfTcTDJiMXWOg" \
     -H "Content-Type: application/json" \
     -d '{
           "toUser": "user_2",
           "amount": "100"
         }'
```
Пример ответа:

``` 200 OK ```


### Получить информацию о монетах, инвентаре и истории транзакций <a name="get-info"></a>
```curl -X GET http://localhost:8080/api/info \
     -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.Cu0jgoMXBW7FiljBNqT63i0TKnwWUFBfTcTDJiMXWOg" \
     -H "Content-Type: application/json"
```

Пример ответа:

```json
{
   { "coinHistory":{"received":[],"sent":[{"amount":100,"toUser":"user_2"}]},"coins":850,"inventory":[{"quantity":1,"type":"book"}]}
}
```

### Unit-тесты

Для тестирования методов бизнес-логики (internal/application) и API (internal/facade) были добавлены модульные табличные тесты. Все зависимости сервисов, такие как application.Service у API и storage.Service у слоя приложения, были описаны через интерфейсы. Это позволило подменять их заглушками, сгенерированными инструментом go.uber.org/mock/mockgen, и настраивать их поведение для тестирования различных сценариев работы методов. Такой подход обеспечил изолированную проверку корректности логики каждого метода.

Для запуска `make unit_test`

Вывод:

```
ok      github.com/azaliaz/avito-shop/internal/application/tests 1.020s
ok      github.com/azaliaz/avito-shop/internal/facade/rest/tests 0.532s
```

# **Дополнительные задания**

### Нагрузочное тестирование 

Проведено нагрузочное тестирование (load_tests) 

Ниже приведены результаты нагрузочного тестирования эндпоинта `api/buy/:item`:
```
✓ is status 200

     checks.........................: 100.00% 16360 out of 16360
     data_received..................: 1.5 MB  42 kB/s
     data_sent......................: 4.0 MB  109 kB/s
     http_req_blocked...............: avg=1.33ms   min=65µs     med=141µs    max=19s     p(90)=250µs  p(95)=293µs 
     http_req_connecting............: avg=1.3ms    min=54µs     med=114µs    max=19s     p(90)=199µs  p(95)=231µs 
   ✓ http_req_duration..............: avg=849.87µs min=640µs    med=763µs    max=22.16ms p(90)=912µs  p(95)=1.1ms 
       { expected_response:true }...: avg=849.87µs min=640µs    med=763µs    max=22.16ms p(90)=912µs  p(95)=1.1ms 
   ✓ http_req_failed................: 0.00%   0 out of 16360
     http_req_receiving.............: avg=31.25µs  min=8µs      med=24µs     max=6.6ms   p(90)=45µs   p(95)=55µs  
     http_req_sending...............: avg=21.91µs  min=4µs      med=18µs     max=3.97ms  p(90)=33µs   p(95)=42µs  
     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s       max=0s      p(90)=0s     p(95)=0s    
     http_req_waiting...............: avg=796.71µs min=590µs    med=718µs    max=22.01ms p(90)=847µs  p(95)=1.02ms
     http_reqs......................: 16360   443.841346/s
     iteration_duration.............: avg=2.24ms   min=792.66µs med=966.62µs max=19.02s  p(90)=1.21ms p(95)=1.46ms
     iterations.....................: 16360   443.841346/s
     vus............................: 1       min=1              max=1
     vus_max........................: 1       min=1              max=1


running (0m36.9s), 0/1 VUs, 16360 complete and 0 interrupted iterations
default ✓ [======================================] 1 VUs  30s
```

Ниже приведены результаты нагрузочного тестирования эндпоинта `api/sendCoin`:

```
✓ is status 200

     checks.........................: 100.00% 16365 out of 16365
     data_received..................: 1.5 MB  40 kB/s
     data_sent......................: 4.9 MB  128 kB/s
     http_req_blocked...............: avg=1.31ms   min=70µs     med=131µs    max=19s     p(90)=218µs  p(95)=271µs 
     http_req_connecting............: avg=1.29ms   min=54µs     med=108µs    max=19s     p(90)=175µs  p(95)=215µs 
   ✓ http_req_duration..............: avg=959.09µs min=679µs    med=804µs    max=45.37ms p(90)=1.13ms p(95)=1.4ms 
       { expected_response:true }...: avg=959.09µs min=679µs    med=804µs    max=45.37ms p(90)=1.13ms p(95)=1.4ms 
   ✓ http_req_failed................: 0.00%   0 out of 16365
     http_req_receiving.............: avg=27.26µs  min=8µs      med=21µs     max=8.07ms  p(90)=42µs   p(95)=53µs  
     http_req_sending...............: avg=19.79µs  min=4µs      med=15µs     max=1.57ms  p(90)=33µs   p(95)=41µs  
     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s       max=0s      p(90)=0s     p(95)=0s    
     http_req_waiting...............: avg=912.03µs min=616µs    med=766µs    max=45.01ms p(90)=1.07ms p(95)=1.33ms
     http_reqs......................: 16365   427.200873/s
     iteration_duration.............: avg=2.33ms   min=811.16µs med=999.08µs max=19.02s  p(90)=1.38ms p(95)=1.74ms
     iterations.....................: 16365   427.200873/s
     vus............................: 1       min=1              max=1
     vus_max........................: 1       min=1              max=1


running (0m38.3s), 0/1 VUs, 16365 complete and 0 interrupted iterations
default ✓ [======================================] 1 VUs  30s
```

### Реализованы интеграционные тесты, для запуска `make integration_tests`

Для тестирования методов слоя репозитория(internal/storage) были добавлены интеграционные тесты с использованием пакета github.com/testcontainers/testcontainers-go. Этот пакет позволил запускать тестовую PostgreSQL базу данных в контейнере, заменяя реальную базу данных в процессе выполнения тестов. Все SQL-запросы выполнялись на тестовой БД, что обеспечивало возможность проверки правильности интеграции базы данных в приложение. Таким образом, удалось проверить выполнение запросов без ошибок, их корректное поведение и взаимодействие с базой данных.

### Описание тестов
* `TestAuth` - тестирует аутентификацию пользователей:
    * `Authenticate_and_create_new_user` — успешная аутентификация и создание нового пользователя.
    * `Authenticate_with_incorrect_password` — попытка аутентификации с неверным паролем.
    * `Authenticate_with_empty_password` — попытка аутентификации с пустым паролем.
    * `Authenticate_multiple_times_with_same_credentials` — аутентификация с одинаковыми данными несколько раз.
* `TestBuyItem` - тестирует покупку товаров.
* `TestGetBalance` - тестирует получение баланса пользователя.
* `TestGetCoinHistory` - тестирует получение истории транзакций с монетами.
* `TestGetInventory` - тестирует получение инвентаря пользователя.
* `TestSendCoin` - тестирует отправку монет:
    * `Test_successful_SendCoin` — успешная отправка монет.
    * `Test_SendCoin_with_insufficient_balance` — попытка отправки монет с недостаточным балансом.
    * `Test_SendCoin_to_non-existing_user` — попытка отправки монет несуществующему пользователю.

### Результаты тестов

```
--- PASS: TestRepositorySuite (4.28s)
    --- PASS: TestRepositorySuite/TestAuth (0.12s)
        --- PASS: TestRepositorySuite/TestAuth/Authenticate_and_create_new_user (0.00s)
        --- PASS: TestRepositorySuite/TestAuth/Authenticate_with_incorrect_password (0.00s)
        --- PASS: TestRepositorySuite/TestAuth/Authenticate_with_empty_password (0.00s)
        --- PASS: TestRepositorySuite/TestAuth/Authenticate_multiple_times_with_same_credentials (0.00s)
    --- PASS: TestRepositorySuite/TestBuyItem (0.11s)
    --- PASS: TestRepositorySuite/TestGetBalance (0.06s)
    --- PASS: TestRepositorySuite/TestGetCoinHistory (0.07s)
    --- PASS: TestRepositorySuite/TestGetInventory (0.07s)
    --- PASS: TestRepositorySuite/TestSendCoin (0.07s)
        --- PASS: TestRepositorySuite/TestSendCoin/Test_successful_SendCoin (0.00s)
        --- PASS: TestRepositorySuite/TestSendCoin/Test_SendCoin_with_insufficient_balance (0.00s)
        --- PASS: TestRepositorySuite/TestSendCoin/Test_SendCoin_to_non-existing_user (0.01s)
PASS
ok      github.com/azaliaz/avito-shop/internal/storage/tests     4.805s
```



### Конфигурация линтера

Конфигурация линтера настроена для тщательной проверки кода с фокусом на ошибки, неиспользуемые переменные, сложность функций и корректность типов. Включены линтеры для проверки ошибок, использования неэффективных присваиваний, а также для улучшения стиля и качества кода. Некоторые линтеры, такие как gosimple, отключены для упрощения анализа. Установлены ограничения по времени выполнения и максимальному количеству ошибок, а также исключены некоторые правила для конкретных файлов.

Для тестирования методов слоя репо были добвлены интеграционные тесты с использованием пакета "github.com/testcontainers/testcontainers-go", благодаря которому при выполнении теста база данных подменялась тестовым контейнером с postgreSQL и все SQL запросы выполнялись на тестовой БД. Это позволило проверить интеграцию БД в приложение, а также корректное выполнение запросов без исключений, их поведение.