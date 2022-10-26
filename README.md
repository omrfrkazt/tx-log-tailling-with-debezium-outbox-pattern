# transaction log tailing pattern w/ Debezium + Kafka & Outbox Design Pattern

> Auth microservice implementation.


https://medium.com/trendyol-tech/transaction-log-tailing-with-debezium-part-1-aeb968d72220


<img src="./dbz.png" alt="Debezium with Kafka">


> PostgreSQL database 

> SQLC
https://github.com/kyleconroy/sqlc

> Go-Kit
https://gokit.io


> Docker-compose testing, main.go will subscribe to the given topics (tables) and print events whenever the DB state changed.

> Microservices will write to own database. Debezium will automatically produce the events.

> Consumer microservices will continue to listen for events.

> In this example, outbox design pattern is used simply.
