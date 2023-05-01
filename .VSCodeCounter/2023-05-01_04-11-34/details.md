# Details

Date : 2023-05-01 04:11:34

Directory /root/go/src/github.com/geraldkohn/im

Total : 50 files,  5344 codes, 0 comments, 686 blanks, all 6030 lines

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)

## Files
| filename | language | code | comment | blank | total |
| :--- | :--- | ---: | ---: | ---: | ---: |
| [deploy/etcd_deployment.yml](/deploy/etcd_deployment.yml) | YAML | 82 | 0 | 1 | 83 |
| [go.mod](/go.mod) | Go Module File | 96 | 0 | 5 | 101 |
| [go.sum](/go.sum) | Go Checksum File | 686 | 0 | 1 | 687 |
| [internal/api/chat/chat.pb.go](/internal/api/chat/chat.pb.go) | Go | 1,150 | 0 | 154 | 1,304 |
| [internal/api/chat/chat.proto](/internal/api/chat/chat.proto) | Protocol Buffers | 98 | 0 | 21 | 119 |
| [internal/api/chat/chat_grpc.pb.go](/internal/api/chat/chat_grpc.pb.go) | Go | 166 | 0 | 26 | 192 |
| [internal/gateway/grpc_server.go](/internal/gateway/grpc_server.go) | Go | 64 | 0 | 11 | 75 |
| [internal/gateway/http_server.go](/internal/gateway/http_server.go) | Go | 40 | 0 | 8 | 48 |
| [internal/gateway/init.go](/internal/gateway/init.go) | Go | 61 | 0 | 12 | 73 |
| [internal/gateway/keep_alive.go](/internal/gateway/keep_alive.go) | Go | 43 | 0 | 7 | 50 |
| [internal/gateway/ws_logic.go](/internal/gateway/ws_logic.go) | Go | 179 | 0 | 18 | 197 |
| [internal/gateway/ws_server.go](/internal/gateway/ws_server.go) | Go | 229 | 0 | 19 | 248 |
| [internal/pusher/grpc_server.go](/internal/pusher/grpc_server.go) | Go | 45 | 0 | 11 | 56 |
| [internal/pusher/init.go](/internal/pusher/init.go) | Go | 90 | 0 | 12 | 102 |
| [internal/pusher/logic.go](/internal/pusher/logic.go) | Go | 38 | 0 | 4 | 42 |
| [internal/transfer/init.go](/internal/transfer/init.go) | Go | 93 | 0 | 11 | 104 |
| [internal/transfer/msg_handler.go](/internal/transfer/msg_handler.go) | Go | 156 | 0 | 8 | 164 |
| [internal/transfer/retry_handler.go](/internal/transfer/retry_handler.go) | Go | 27 | 0 | 4 | 31 |
| [pkg/common/constant/defs.go](/pkg/common/constant/defs.go) | Go | 26 | 0 | 6 | 32 |
| [pkg/common/constant/errors.go](/pkg/common/constant/errors.go) | Go | 45 | 0 | 8 | 53 |
| [pkg/common/cronjob/job.go](/pkg/common/cronjob/job.go) | Go | 35 | 0 | 10 | 45 |
| [pkg/common/cronjob/job_test.go](/pkg/common/cronjob/job_test.go) | Go | 12 | 0 | 3 | 15 |
| [pkg/common/cronjob/scheduler.go](/pkg/common/cronjob/scheduler.go) | Go | 89 | 0 | 15 | 104 |
| [pkg/common/cronjob/scheduler_test.go](/pkg/common/cronjob/scheduler_test.go) | Go | 30 | 0 | 6 | 36 |
| [pkg/common/db/db.go](/pkg/common/db/db.go) | Go | 72 | 0 | 13 | 85 |
| [pkg/common/db/mongo_model.go](/pkg/common/db/mongo_model.go) | Go | 70 | 0 | 9 | 79 |
| [pkg/common/db/mysql_group_model.go](/pkg/common/db/mysql_group_model.go) | Go | 33 | 0 | 9 | 42 |
| [pkg/common/db/mysql_model.go](/pkg/common/db/mysql_model.go) | Go | 83 | 0 | 11 | 94 |
| [pkg/common/db/mysql_user_model.go](/pkg/common/db/mysql_user_model.go) | Go | 21 | 0 | 6 | 27 |
| [pkg/common/db/redis_model.go](/pkg/common/db/redis_model.go) | Go | 85 | 0 | 13 | 98 |
| [pkg/common/grpc-etcdv3/pool.go](/pkg/common/grpc-etcdv3/pool.go) | Go | 225 | 0 | 30 | 255 |
| [pkg/common/grpc-etcdv3/register.go](/pkg/common/grpc-etcdv3/register.go) | Go | 80 | 0 | 17 | 97 |
| [pkg/common/grpc-etcdv3/resolver.go](/pkg/common/grpc-etcdv3/resolver.go) | Go | 226 | 0 | 40 | 266 |
| [pkg/common/http/http_client.go](/pkg/common/http/http_client.go) | Go | 42 | 0 | 6 | 48 |
| [pkg/common/kafka/consumer.go](/pkg/common/kafka/consumer.go) | Go | 35 | 0 | 8 | 43 |
| [pkg/common/kafka/consumer_group.go](/pkg/common/kafka/consumer_group.go) | Go | 47 | 0 | 9 | 56 |
| [pkg/common/kafka/kafka_test.go](/pkg/common/kafka/kafka_test.go) | Go | 211 | 0 | 36 | 247 |
| [pkg/common/kafka/producer.go](/pkg/common/kafka/producer.go) | Go | 51 | 0 | 6 | 57 |
| [pkg/common/logger/logger.go](/pkg/common/logger/logger.go) | Go | 44 | 0 | 11 | 55 |
| [pkg/common/service-discovery/etcd_register_discovery.go](/pkg/common/service-discovery/etcd_register_discovery.go) | Go | 165 | 0 | 19 | 184 |
| [pkg/common/service-discovery/etcd_test.go](/pkg/common/service-discovery/etcd_test.go) | Go | 21 | 0 | 4 | 25 |
| [pkg/common/service-discovery/interface.go](/pkg/common/service-discovery/interface.go) | Go | 11 | 0 | 3 | 14 |
| [pkg/common/setting/setting.go](/pkg/common/setting/setting.go) | Go | 59 | 0 | 11 | 70 |
| [pkg/utils/cors_middleware.go](/pkg/utils/cors_middleware.go) | Go | 21 | 0 | 4 | 25 |
| [pkg/utils/cors_middleware_test.go](/pkg/utils/cors_middleware_test.go) | Go | 58 | 0 | 10 | 68 |
| [pkg/utils/jwt_token.go](/pkg/utils/jwt_token.go) | Go | 34 | 0 | 10 | 44 |
| [pkg/utils/jwt_token_test.go](/pkg/utils/jwt_token_test.go) | Go | 26 | 0 | 5 | 31 |
| [pkg/utils/md5.go](/pkg/utils/md5.go) | Go | 11 | 0 | 3 | 14 |
| [pkg/utils/md5_test.go](/pkg/utils/md5_test.go) | Go | 11 | 0 | 5 | 16 |
| [pkg/utils/time_formatter.go](/pkg/utils/time_formatter.go) | Go | 22 | 0 | 7 | 29 |

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)