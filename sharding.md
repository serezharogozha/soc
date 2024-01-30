# Шардинг

Шардирование реалищовано посредством использования citus, который позволяет распределить данные по разным нодам.
Все необходимые изменения произведены в файле `docker-compose.yml` в разделе `citus`.
А именно:
- добавлен coordinator и 2 worker

Для того, чтобы проверить работу шардинга, необходимо:
- запустить контейнеры
- подключиться к контейнеру с postgresql
- ```
  SELECT master_add_node('citus_worker_1', 5432);
  SELECT master_add_node('citus_worker_2', 5432);
  ```
- выполнить команду ```SELECT create_distributed_table('messages', 'dialogue_id');```
- выполнив команду ```SELECT master_get_active_worker_nodes(); ``` мы видим два активных воркера
- далее нужно нагенерировать данные в таблиц dialogues и messages
```
INSERT INTO dialogues(from_user_id, to_user_id)
    SELECT
      FLOOR(RANDOM()*(100-1+1))+1, 
      FLOOR(random()*(200-100+1))+100 
    FROM generate_series(1, 50000);

INSERT INTO messages(dialogue_id, from_user_id, to_user_id, text)
SELECT
    d.id,
    d.from_user_id,
    d.to_user_id,
    md5(random()::text)
FROM
    (SELECT *, ROW_NUMBER() OVER () as rownum FROM dialogues) d
        CROSS JOIN
    generate_series(1, CEIL(300000.0 / (SELECT COUNT(*) FROM dialogues))) as g
WHERE
        d.rownum <= 300000;
```
- выполнив команду ```EXPLAIN SELECT * FROM messages WHERE from_user_id = ? limit 10;``` 
посдтавляя разные значения `from_user_id` мы видим, что данные распределены по двум воркерам
- добавим еще пару воркеров в docker-compose скопирова предыдущий 
-   ```
    SELECT master_add_node('citus_worker_3', 5432);
    SELECT master_add_node('citus_worker_4', 5432);
    ```
- после этого выполним команду ```SELECT master_get_active_worker_nodes(); ``` и увидим, что воркеров стало 4
- ```
  alter system set wal_level = logical;
  SELECT run_command_on_workers('alter system set wal_level = logical');
  ```
  после этого перезапустим контейнеры
- если мы не можем использовать primary Или unique key, то можно использовать 
```ALTER TABLE dialogues REPLICA IDENTITY FULL;```
    будут включены все колонки, надо быть осторожным, если колонок много, то это может сильно увеличить размер и время репликации
- теперь нам нужно перераспределить данные по воркерам, для этого выполним команду ```SELECT rebalance_table_shards('dialogues');```
- следить за выполнение ```SELECT * FROM citus_rebalance_status();```
- проверим, что данные равномерно распределены по всем шардам 
```
SELECT nodename, count(*)
  FROM citus_shards GROUP BY nodename;
```
