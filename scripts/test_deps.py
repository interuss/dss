import copy
# from os import path

parents = {
    'cloud.google.com/go': {
        'github.com/golang-migrate/migrate/v4',
        'google.golang.org/api',
        'github.com/interuss/dss',
        'cloud.google.com/go/storage',
        'cloud.google.com/go/bigquery',
        'cloud.google.com/go/spanner',
        'golang.org/x/oauth2',
        'cloud.google.com/go/pubsub',
        'google.golang.org/grpc',
        'cloud.google.com/go/datastore'
    },
    'github.com/cockroachdb/cockroach-go': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/interuss/dss'
    },
    'github.com/coreos/go-semver': {
        'github.com/interuss/dss'
    },
    'github.com/golang-jwt/jwt': {
        'github.com/interuss/dss'
    },
    'github.com/golang-migrate/migrate/v4': {
        'github.com/interuss/dss'
    },
    'github.com/golang/geo': {
        'github.com/interuss/dss'
    },
    'github.com/golang/protobuf': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'google.golang.org/protobuf',
        'github.com/prometheus/client_model',
        'github.com/ktrysmt/go-bitbucket',
        'cloud.google.com/go/bigquery',
        'go.opencensus.io',
        'github.com/cncf/udpa/go',
        'github.com/grpc-ecosystem/grpc-gateway',
        'google.golang.org/api',
        'github.com/envoyproxy/go-control-plane',
        'cloud.google.com/go/pubsub',
        'google.golang.org/appengine',
        'github.com/interuss/dss',
        'cloud.google.com/go/spanner',
        'github.com/dhui/dktest',
        'cloud.google.com/go/storage',
        'google.golang.org/genproto',
        'github.com/onsi/gomega',
        'google.golang.org/grpc',
        'cloud.google.com/go/datastore',
        'cloud.google.com/go'
    },
    'github.com/google/uuid': {
        'github.com/snowflakedb/gosnowflake',
        'google.golang.org/grpc',
        'github.com/interuss/dss'
    },
    'github.com/grpc-ecosystem/go-grpc-middleware': {
        'github.com/interuss/dss'
    },
    'github.com/grpc-ecosystem/grpc-gateway': {
        'github.com/interuss/dss'
    },
    'github.com/interuss/stacktrace': {
        'github.com/interuss/dss'
    },
    'github.com/jonboulle/clockwork': {
        'github.com/interuss/dss'
    },
    'github.com/lib/pq': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/dhui/dktest',
        'github.com/jackc/pgx/v4',
        'github.com/interuss/dss',
        'github.com/jmoiron/sqlx',
        'github.com/jackc/pgtype'
    },
    'github.com/pkg/errors': {
        'github.com/jackc/pgconn',
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'github.com/dhui/dktest',
        'github.com/jackc/pgx/v4',
        'github.com/neo4j/neo4j-go-driver',
        'github.com/interuss/dss',
        'go.uber.org/zap',
        'github.com/jackc/pgproto3',
        'github.com/Microsoft/go-winio',
        'github.com/rs/zerolog',
        'github.com/jackc/pgproto3/v2'
    },
    'github.com/robfig/cron/v3': {
        'github.com/interuss/dss'
    },
    'github.com/stretchr/testify': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'github.com/apache/arrow/go/arrow',
        'go.opencensus.io',
        'github.com/gobuffalo/here',
        'go.uber.org/atomic',
        'go.uber.org/zap',
        'github.com/jackc/pgmock',
        'github.com/envoyproxy/go-control-plane',
        'go.uber.org/multierr',
        'github.com/markbates/pkger',
        'github.com/jackc/pgconn',
        'github.com/jackc/pgx/v4',
        'github.com/interuss/dss',
        'github.com/sirupsen/logrus',
        'github.com/mutecomm/go-sqlcipher/v4',
        'github.com/jackc/puddle',
        'github.com/ClickHouse/clickhouse-go',
        'github.com/dhui/dktest',
        'github.com/stretchr/objx',
        'github.com/jackc/pgtype',
        'github.com/jackc/pgpassfile',
        'github.com/jackc/pgproto3/v2'
    },
    'go.uber.org/zap': {
        'github.com/jackc/pgx/v4',
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'github.com/interuss/dss'
    },
    'google.golang.org/genproto': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'google.golang.org/protobuf',
        'github.com/grpc-ecosystem/grpc-gateway',
        'github.com/dhui/dktest',
        'google.golang.org/api',
        'github.com/interuss/dss',
        'cloud.google.com/go/storage',
        'cloud.google.com/go/bigquery',
        'go.opencensus.io',
        'github.com/envoyproxy/go-control-plane',
        'cloud.google.com/go/spanner',
        'cloud.google.com/go/pubsub',
        'google.golang.org/grpc',
        'cloud.google.com/go/datastore',
        'cloud.google.com/go'
    },
    'google.golang.org/grpc': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'github.com/grpc-ecosystem/grpc-gateway',
        'github.com/dhui/dktest',
        'google.golang.org/api',
        'github.com/interuss/dss',
        'cloud.google.com/go/storage',
        'cloud.google.com/go/bigquery',
        'google.golang.org/genproto',
        'go.opencensus.io',
        'github.com/cncf/udpa/go',
        'cloud.google.com/go/spanner',
        'github.com/envoyproxy/go-control-plane',
        'github.com/googleapis/gax-go/v2',
        'cloud.google.com/go/pubsub',
        'cloud.google.com/go/datastore',
        'cloud.google.com/go'
    },
    'google.golang.org/protobuf': {
        'github.com/dhui/dktest',
        'github.com/interuss/dss',
        'google.golang.org/genproto',
        'github.com/envoyproxy/go-control-plane',
        'github.com/golang/protobuf',
        'google.golang.org/grpc',
        'cloud.google.com/go'
    },
    'gopkg.in/square/go-jose.v2': {
        'github.com/interuss/dss'
    },
    'cloud.google.com/go/storage': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/fsouza/fake-gcs-server',
        'cloud.google.com/go/bigquery',
        'cloud.google.com/go/pubsub',
        'cloud.google.com/go'
    },
    'github.com/golang/mock': {
        'google.golang.org/grpc',
        'github.com/neo4j/neo4j-go-driver',
        'cloud.google.com/go'
    },
    'github.com/google/go-cmp': {
        'google.golang.org/protobuf',
        'github.com/fsouza/fake-gcs-server',
        'google.golang.org/api',
        'cloud.google.com/go/storage',
        'cloud.google.com/go/bigquery',
        'go.opencensus.io',
        'github.com/envoyproxy/go-control-plane',
        'cloud.google.com/go/spanner',
        'github.com/golang/protobuf',
        'cloud.google.com/go/pubsub',
        'google.golang.org/grpc',
        'cloud.google.com/go/datastore',
        'cloud.google.com/go'
    },
    'github.com/google/martian/v3': {
        'cloud.google.com/go'
    },
    'github.com/google/pprof': {
        'cloud.google.com/go'
    },
    'github.com/googleapis/gax-go/v2': {
        'google.golang.org/api',
        'cloud.google.com/go/storage',
        'cloud.google.com/go/bigquery',
        'cloud.google.com/go/spanner',
        'cloud.google.com/go/pubsub',
        'cloud.google.com/go/datastore',
        'cloud.google.com/go'
    },
    'github.com/jstemmer/go-junit-report': {
        'cloud.google.com/go/storage',
        'cloud.google.com/go'
    },
    'go.opencensus.io': {
        'google.golang.org/api',
        'cloud.google.com/go/storage',
        'cloud.google.com/go/spanner',
        'cloud.google.com/go/pubsub',
        'cloud.google.com/go'
    },
    'golang.org/x/lint': {
        'google.golang.org/api',
        'go.uber.org/zap',
        'cloud.google.com/go/bigquery',
        'google.golang.org/genproto',
        'go.uber.org/multierr',
        'cloud.google.com/go',
        'cloud.google.com/go/pubsub',
        'google.golang.org/grpc',
        'go.uber.org/atomic'
    },
    'golang.org/x/net': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'github.com/ktrysmt/go-bitbucket',
        'cloud.google.com/go/bigquery',
        'go.opencensus.io',
        'golang.org/x/oauth2',
        'golang.org/x/crypto',
        'github.com/grpc-ecosystem/grpc-gateway',
        'google.golang.org/api',
        'github.com/xanzy/go-gitlab',
        'cloud.google.com/go/pubsub',
        'golang.org/x/tools',
        'google.golang.org/appengine',
        'github.com/jackc/pgx/v4',
        'github.com/mutecomm/go-sqlcipher/v4',
        'github.com/dhui/dktest',
        'cloud.google.com/go/storage',
        'google.golang.org/genproto',
        'github.com/google/martian/v3',
        'github.com/onsi/gomega',
        'google.golang.org/grpc',
        'cloud.google.com/go'
    },
    'golang.org/x/oauth2': {
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'github.com/grpc-ecosystem/grpc-gateway',
        'google.golang.org/api',
        'cloud.google.com/go/storage',
        'github.com/ktrysmt/go-bitbucket',
        'github.com/xanzy/go-gitlab',
        'cloud.google.com/go/pubsub',
        'google.golang.org/grpc',
        'cloud.google.com/go'
    },
    'golang.org/x/text': {
        'github.com/jackc/pgconn',
        'golang.org/x/net',
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'google.golang.org/appengine',
        'github.com/dhui/dktest',
        'google.golang.org/api',
        'github.com/jackc/pgx/v4',
        'rsc.io/sampler',
        'go.opencensus.io',
        'golang.org/x/image',
        'github.com/onsi/gomega',
        'google.golang.org/grpc',
        'cloud.google.com/go'
    },
    'golang.org/x/tools': {
        'github.com/golang-migrate/migrate/v4',
        'honnef.co/go/tools',
        'cloud.google.com/go/bigquery',
        'golang.org/x/mod',
        'go.uber.org/atomic',
        'google.golang.org/api',
        'github.com/kisielk/errcheck',
        'go.uber.org/multierr',
        'cloud.google.com/go/pubsub',
        'golang.org/x/text',
        'google.golang.org/appengine',
        'github.com/jackc/pgx/v4',
        'cloud.google.com/go/spanner',
        'golang.org/x/exp',
        'github.com/golang/mock',
        'cloud.google.com/go/storage',
        'google.golang.org/genproto',
        'golang.org/x/lint',
        'github.com/rs/zerolog',
        'google.golang.org/grpc',
        'cloud.google.com/go/datastore',
        'cloud.google.com/go'
    },
    'golang.org/x/xerrors': {
        'github.com/jackc/pgconn',
        'golang.org/x/tools',
        'github.com/apache/arrow/go/arrow',
        'github.com/grpc-ecosystem/grpc-gateway',
        'github.com/google/go-cmp',
        'github.com/jackc/pgx/v4',
        'github.com/jackc/pgmock',
        'github.com/jackc/pgtype',
        'cloud.google.com/go/spanner',
        'github.com/onsi/gomega',
        'golang.org/x/exp',
        'golang.org/x/mod',
        'cloud.google.com/go'
    },
    'google.golang.org/api': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/fsouza/fake-gcs-server',
        'cloud.google.com/go/storage',
        'cloud.google.com/go/bigquery',
        'cloud.google.com/go/spanner',
        'cloud.google.com/go/pubsub',
        'cloud.google.com/go/datastore',
        'cloud.google.com/go'
    },
    'cloud.google.com/go/spanner': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/ClickHouse/clickhouse-go': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/aws/aws-sdk-go': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/bitly/go-hostpool': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/bmizerany/assert': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/cenkalti/backoff/v4': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/containerd/containerd': {
        'github.com/dhui/dktest',
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/cznic/mathutil': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/denisenkom/go-mssqldb': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/dhui/dktest': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/docker/docker': {
        'github.com/dhui/dktest',
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/edsrzf/mmap-go': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/fsouza/fake-gcs-server': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/go-sql-driver/mysql': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/jmoiron/sqlx'
    },
    'github.com/gobuffalo/here': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/markbates/pkger'
    },
    'github.com/gocql/gocql': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/golang/snappy': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/gocql/gocql'
    },
    'github.com/google/go-github': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/gorilla/mux': {
        'github.com/fsouza/fake-gcs-server',
        'github.com/dhui/dktest',
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/hashicorp/go-multierror': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/jackc/pgconn': {
        'github.com/jackc/pgx/v4',
        'github.com/golang-migrate/migrate/v4',
        'github.com/jackc/pgmock'
    },
    'github.com/kardianos/osext': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/ktrysmt/go-bitbucket': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/markbates/pkger': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/mattn/go-sqlite3': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/jmoiron/sqlx'
    },
    'github.com/mutecomm/go-sqlcipher/v4': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/nakagami/firebirdsql': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/neo4j/neo4j-go-driver': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/sirupsen/logrus': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'github.com/fsouza/fake-gcs-server',
        'github.com/dhui/dktest',
        'github.com/jackc/pgx/v4',
        'github.com/Microsoft/go-winio'
    },
    'github.com/snowflakedb/gosnowflake': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/tidwall/pretty': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/xanzy/go-gitlab': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/xdg/scram': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/xdg/stringprep': {
        'github.com/golang-migrate/migrate/v4'
    },
    'gitlab.com/nyarla/go-crypt': {
        'github.com/golang-migrate/migrate/v4'
    },
    'go.mongodb.org/mongo-driver': {
        'github.com/golang-migrate/migrate/v4'
    },
    'go.uber.org/atomic': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'github.com/jackc/pgx/v4',
        'go.uber.org/zap',
        'go.uber.org/multierr'
    },
    'golang.org/x/crypto': {
        'github.com/jackc/pgconn',
        'golang.org/x/net',
        'github.com/golang-migrate/migrate/v4',
        'google.golang.org/appengine',
        'github.com/snowflakedb/gosnowflake',
        'github.com/jackc/pgx/v4',
        'github.com/denisenkom/go-mssqldb',
        'golang.org/x/mod'
    },
    'golang.org/x/sys': {
        'github.com/golang-migrate/migrate/v4',
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'github.com/google/pprof',
        'github.com/ktrysmt/go-bitbucket',
        'cloud.google.com/go/bigquery',
        'go.opencensus.io',
        'golang.org/x/crypto',
        'google.golang.org/api',
        'github.com/onsi/ginkgo',
        'cloud.google.com/go/pubsub',
        'golang.org/x/net',
        'google.golang.org/appengine',
        'github.com/jackc/pgx/v4',
        'golang.org/x/mobile',
        'github.com/sirupsen/logrus',
        'github.com/mattn/go-isatty',
        'golang.org/x/exp',
        'github.com/dhui/dktest',
        'cloud.google.com/go/storage',
        'github.com/Microsoft/go-winio',
        'github.com/onsi/gomega',
        'google.golang.org/grpc',
        'cloud.google.com/go/datastore',
        'cloud.google.com/go'
    },
    'golang.org/x/time': {
        'github.com/dhui/dktest',
        'github.com/golang-migrate/migrate/v4',
        'cloud.google.com/go',
        'cloud.google.com/go/pubsub'
    },
    'modernc.org/b': {
        'github.com/golang-migrate/migrate/v4'
    },
    'modernc.org/db': {
        'github.com/golang-migrate/migrate/v4'
    },
    'modernc.org/file': {
        'github.com/golang-migrate/migrate/v4'
    },
    'modernc.org/fileutil': {
        'github.com/golang-migrate/migrate/v4'
    },
    'modernc.org/golex': {
        'github.com/golang-migrate/migrate/v4'
    },
    'modernc.org/internal': {
        'github.com/golang-migrate/migrate/v4'
    },
    'modernc.org/lldb': {
        'github.com/golang-migrate/migrate/v4'
    },
    'modernc.org/ql': {
        'github.com/golang-migrate/migrate/v4'
    },
    'modernc.org/sortutil': {
        'github.com/golang-migrate/migrate/v4'
    },
    'modernc.org/strutil': {
        'github.com/golang-migrate/migrate/v4'
    },
    'modernc.org/zappy': {
        'github.com/golang-migrate/migrate/v4'
    },
    'github.com/go-kit/kit': {
        'github.com/grpc-ecosystem/go-grpc-middleware'
    },
    'github.com/go-logfmt/logfmt': {
        'github.com/grpc-ecosystem/go-grpc-middleware'
    },
    'github.com/go-stack/stack': {
        'github.com/jackc/pgx/v4',
        'github.com/grpc-ecosystem/go-grpc-middleware'
    },
    'github.com/gogo/protobuf': {
        'github.com/dhui/dktest',
        'github.com/grpc-ecosystem/go-grpc-middleware'
    },
    'github.com/opentracing/opentracing-go': {
        'github.com/grpc-ecosystem/go-grpc-middleware'
    },
    'go.uber.org/multierr': {
        'github.com/jackc/pgtype',
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'github.com/jackc/pgx/v4',
        'go.uber.org/zap'
    },
    'golang.org/x/exp': {
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'golang.org/x/mobile',
        'cloud.google.com/go/storage',
        'cloud.google.com/go/bigquery',
        'google.golang.org/genproto',
        'cloud.google.com/go/pubsub',
        'cloud.google.com/go/datastore',
        'cloud.google.com/go'
    },
    'github.com/antihax/optional': {
        'github.com/grpc-ecosystem/grpc-gateway'
    },
    'github.com/ghodss/yaml': {
        'github.com/grpc-ecosystem/grpc-gateway'
    },
    'github.com/golang/glog': {
        'google.golang.org/grpc',
        'github.com/grpc-ecosystem/grpc-gateway'
    },
    'github.com/rogpeppe/fastuuid': {
        'github.com/grpc-ecosystem/grpc-gateway'
    },
    'gopkg.in/yaml.v2': {
        'github.com/stretchr/testify',
        'github.com/grpc-ecosystem/grpc-gateway',
        'github.com/dhui/dktest',
        'go.uber.org/zap',
        'github.com/gobuffalo/here',
        'github.com/onsi/gomega'
    },
    'github.com/davecgh/go-spew': {
        'github.com/apache/arrow/go/arrow',
        'github.com/stretchr/testify',
        'github.com/stretchr/objx',
        'github.com/sirupsen/logrus',
        'github.com/gobuffalo/here',
        'github.com/mutecomm/go-sqlcipher/v4',
        'go.uber.org/atomic'
    },
    'github.com/pmezard/go-difflib': {
        'github.com/apache/arrow/go/arrow',
        'github.com/sirupsen/logrus',
        'github.com/stretchr/testify'
    },
    'github.com/stretchr/objx': {
        'github.com/jackc/pgx/v4',
        'github.com/sirupsen/logrus',
        'github.com/stretchr/testify'
    },
    'gopkg.in/yaml.v3': {
        'github.com/stretchr/testify'
    },
    'honnef.co/go/tools': {
        'google.golang.org/api',
        'go.uber.org/zap',
        'cloud.google.com/go/bigquery',
        'google.golang.org/genproto',
        'cloud.google.com/go/storage',
        'go.uber.org/multierr',
        'google.golang.org/grpc',
        'cloud.google.com/go'
    },
    'github.com/cncf/udpa/go': {
        'github.com/envoyproxy/go-control-plane',
        'google.golang.org/grpc'
    },
    'github.com/envoyproxy/go-control-plane': {
        'google.golang.org/grpc'
    },
    'cloud.google.com/go/bigquery': {
        'cloud.google.com/go/pubsub',
        'cloud.google.com/go/storage',
        'cloud.google.com/go'
    },
    'github.com/chzyer/logex': {
        'github.com/google/pprof'
    },
    'github.com/chzyer/readline': {
        'github.com/google/pprof'
    },
    'github.com/chzyer/test': {
        'github.com/google/pprof'
    },
    'github.com/ianlancetaylor/demangle': {
        'github.com/google/pprof'
    },
    'github.com/golang/groupcache': {
        'go.opencensus.io',
        'cloud.google.com/go/storage',
        'cloud.google.com/go/bigquery',
        'cloud.google.com/go'
    },
    'golang.org/x/sync': {
        'golang.org/x/tools',
        'github.com/prometheus/client_model',
        'google.golang.org/grpc',
        'google.golang.org/api',
        'github.com/ktrysmt/go-bitbucket',
        'github.com/xanzy/go-gitlab',
        'google.golang.org/genproto',
        'golang.org/x/oauth2',
        'cloud.google.com/go/pubsub',
        'github.com/onsi/gomega',
        'golang.org/x/exp',
        'cloud.google.com/go'
    },
    'google.golang.org/appengine': {
        'golang.org/x/tools',
        'google.golang.org/api',
        'github.com/ktrysmt/go-bitbucket',
        'cloud.google.com/go/storage',
        'github.com/xanzy/go-gitlab',
        'cloud.google.com/go/bigquery',
        'golang.org/x/oauth2',
        'google.golang.org/grpc',
        'cloud.google.com/go/datastore',
        'cloud.google.com/go'
    },
    'github.com/yuin/goldmark': {
        'golang.org/x/tools'
    },
    'golang.org/x/mod': {
        'golang.org/x/tools',
        'honnef.co/go/tools',
        'cloud.google.com/go/bigquery',
        'cloud.google.com/go/pubsub',
        'golang.org/x/exp',
        'cloud.google.com/go'
    },
    'github.com/bkaradzic/go-lz4': {
        'github.com/ClickHouse/clickhouse-go'
    },
    'github.com/cloudflare/golz4': {
        'github.com/ClickHouse/clickhouse-go'
    },
    'github.com/jmoiron/sqlx': {
        'github.com/ClickHouse/clickhouse-go'
    },
    'github.com/pierrec/lz4': {
        'github.com/ClickHouse/clickhouse-go'
    },
    'github.com/jmespath/go-jmespath': {
        'github.com/aws/aws-sdk-go'
    },
    'github.com/golang-sql/civil': {
        'github.com/denisenkom/go-mssqldb'
    },
    'github.com/Azure/go-ansiterm': {
        'github.com/dhui/dktest'
    },
    'github.com/Microsoft/go-winio': {
        'github.com/dhui/dktest'
    },
    'github.com/docker/distribution': {
        'github.com/dhui/dktest'
    },
    'github.com/docker/go-connections': {
        'github.com/dhui/dktest'
    },
    'github.com/docker/go-units': {
        'github.com/dhui/dktest'
    },
    'github.com/gorilla/context': {
        'github.com/dhui/dktest'
    },
    'github.com/kr/pretty': {
        'gopkg.in/errgo.v2',
        'github.com/dhui/dktest',
        'github.com/jackc/pgx/v4',
        'github.com/jackc/pgmock',
        'github.com/gobuffalo/here',
        'github.com/jackc/pgtype'
    },
    'github.com/morikuni/aec': {
        'github.com/dhui/dktest'
    },
    'github.com/opencontainers/go-digest': {
        'github.com/dhui/dktest'
    },
    'github.com/opencontainers/image-spec': {
        'github.com/dhui/dktest'
    },
    'gopkg.in/check.v1': {
        'gopkg.in/errgo.v2',
        'github.com/dhui/dktest',
        'github.com/jackc/pgx/v4',
        'github.com/jackc/pgmock',
        'github.com/gobuffalo/here',
        'github.com/jackc/pgtype',
        'gopkg.in/yaml.v2',
        'gopkg.in/yaml.v3'
    },
    'gotest.tools': {
        'github.com/dhui/dktest'
    },
    'github.com/gorilla/handlers': {
        'github.com/fsouza/fake-gcs-server'
    },
    'github.com/hailocab/go-hostpool': {
        'github.com/gocql/gocql'
    },
    'gopkg.in/inf.v0': {
        'github.com/gocql/gocql'
    },
    'github.com/hashicorp/errwrap': {
        'github.com/hashicorp/go-multierror'
    },
    'github.com/jackc/chunkreader/v2': {
        'github.com/jackc/pgconn',
        'github.com/jackc/pgproto3/v2'
    },
    'github.com/jackc/pgio': {
        'github.com/jackc/pgconn',
        'github.com/jackc/pgx/v4',
        'github.com/jackc/pgproto3',
        'github.com/jackc/pgtype',
        'github.com/jackc/pgproto3/v2'
    },
    'github.com/jackc/pgmock': {
        'github.com/jackc/pgconn'
    },
    'github.com/jackc/pgpassfile': {
        'github.com/jackc/pgconn'
    },
    'github.com/jackc/pgproto3/v2': {
        'github.com/jackc/pgconn',
        'github.com/jackc/pgx/v4',
        'github.com/jackc/pgmock'
    },
    'github.com/k0kubun/colorstring': {
        'github.com/ktrysmt/go-bitbucket'
    },
    'github.com/k0kubun/pp': {
        'github.com/ktrysmt/go-bitbucket'
    },
    'github.com/mattn/go-colorable': {
        'github.com/jackc/pgx/v4',
        'github.com/ktrysmt/go-bitbucket'
    },
    'github.com/mattn/go-isatty': {
        'github.com/jackc/pgx/v4',
        'github.com/mattn/go-colorable',
        'github.com/ktrysmt/go-bitbucket'
    },
    'github.com/mitchellh/mapstructure': {
        'github.com/ktrysmt/go-bitbucket'
    },
    'github.com/onsi/ginkgo': {
        'github.com/onsi/gomega',
        'github.com/neo4j/neo4j-go-driver'
    },
    'github.com/onsi/gomega': {
        'github.com/neo4j/neo4j-go-driver',
        'github.com/onsi/ginkgo'
    },
    'github.com/apache/arrow/go/arrow': {
        'github.com/snowflakedb/gosnowflake'
    },
    'github.com/dgrijalva/jwt-go': {
        'github.com/snowflakedb/gosnowflake'
    },
    'github.com/pkg/browser': {
        'github.com/snowflakedb/gosnowflake'
    },
    'github.com/snowflakedb/glog': {
        'github.com/snowflakedb/gosnowflake'
    },
    'github.com/google/go-querystring': {
        'github.com/xanzy/go-gitlab'
    },
    'github.com/remyoudompheng/bigfft': {
        'modernc.org/sortutil'
    },
    'modernc.org/mathutil': {
        'modernc.org/sortutil'
    },
    'github.com/kr/logfmt': {
        'github.com/go-logfmt/logfmt'
    },
    'github.com/kisielk/errcheck': {
        'github.com/gogo/protobuf'
    },
    'github.com/konsorten/go-windows-terminal-sequences': {
        'github.com/jackc/pgx/v4',
        'github.com/sirupsen/logrus'
    },
    'dmitri.shuralyov.com/gpu/mtl': {
        'golang.org/x/exp'
    },
    'github.com/BurntSushi/xgb': {
        'golang.org/x/exp'
    },
    'github.com/go-gl/glfw/v3.3/glfw': {
        'golang.org/x/exp'
    },
    'golang.org/x/image': {
        'golang.org/x/mobile',
        'golang.org/x/exp'
    },
    'golang.org/x/mobile': {
        'golang.org/x/exp'
    },
    'go.uber.org/tools': {
        'go.uber.org/multierr'
    },
    'github.com/BurntSushi/toml': {
        'google.golang.org/grpc',
        'honnef.co/go/tools'
    },
    'github.com/google/renameio': {
        'honnef.co/go/tools'
    },
    'github.com/kisielk/gotool': {
        'github.com/gogo/protobuf',
        'github.com/kisielk/errcheck',
        'honnef.co/go/tools'
    },
    'github.com/rogpeppe/go-internal': {
        'honnef.co/go/tools'
    },
    'github.com/envoyproxy/protoc-gen-validate': {
        'github.com/cncf/udpa/go',
        'github.com/envoyproxy/go-control-plane',
        'google.golang.org/grpc'
    },
    'github.com/census-instrumentation/opencensus-proto': {
        'github.com/envoyproxy/go-control-plane'
    },
    'github.com/prometheus/client_model': {
        'github.com/envoyproxy/go-control-plane'
    },
    'cloud.google.com/go/datastore': {
        'cloud.google.com/go/storage',
        'cloud.google.com/go'
    },
    'cloud.google.com/go/pubsub': {
        'cloud.google.com/go/bigquery',
        'cloud.google.com/go/storage',
        'cloud.google.com/go/datastore',
        'cloud.google.com/go'
    },
    'github.com/google/martian': {
        'cloud.google.com/go'
    },
    'github.com/client9/misspell': {
        'google.golang.org/grpc'
    },
    'github.com/kr/text': {
        'github.com/kr/pretty'
    },
    'github.com/hashicorp/golang-lru': {
        'go.opencensus.io',
        'google.golang.org/api'
    },
    'github.com/jackc/pgtype': {
        'github.com/jackc/pgx/v4',
        'github.com/jackc/pgmock'
    },
    'rsc.io/quote/v3': {
        'github.com/golang/mock'
    },
    'github.com/hpcloud/tail': {
        'github.com/onsi/gomega',
        'github.com/onsi/ginkgo'
    },
    'github.com/fsnotify/fsnotify': {
        'github.com/onsi/gomega'
    },
    'gopkg.in/fsnotify.v1': {
        'github.com/onsi/gomega'
    },
    'gopkg.in/tomb.v1': {
        'github.com/onsi/gomega'
    },
    'github.com/google/flatbuffers': {
        'github.com/apache/arrow/go/arrow'
    },
    'gopkg.in/errgo.v2': {
        'github.com/rogpeppe/go-internal'
    },
    'github.com/kr/pty': {
        'github.com/jackc/pgx/v4',
        'github.com/kr/text'
    },
    'github.com/google/btree': {
        'cloud.google.com/go'
    },
    'github.com/jackc/pgx/v4': {
        'github.com/jackc/pgtype'
    },
    'github.com/satori/go.uuid': {
        'github.com/jackc/pgtype',
        'github.com/jackc/pgx/v4'
    },
    'github.com/shopspring/decimal': {
        'github.com/jackc/pgtype',
        'github.com/jackc/pgx/v4'
    },
    'rsc.io/sampler': {
        'rsc.io/quote/v3'
    },
    'github.com/go-gl/glfw': {
        'golang.org/x/exp'
    },
    'github.com/cockroachdb/apd': {
        'github.com/jackc/pgx/v4'
    },
    'github.com/coreos/go-systemd': {
        'github.com/jackc/pgx/v4',
        'github.com/rs/zerolog'
    },
    'github.com/jackc/puddle': {
        'github.com/jackc/pgx/v4'
    },
    'github.com/rs/zerolog': {
        'github.com/jackc/pgx/v4'
    },
    'gopkg.in/inconshreveable/log15.v2': {
        'github.com/jackc/pgx/v4'
    },
    'rsc.io/binaryregexp': {
        'cloud.google.com/go'
    },
    'github.com/jackc/chunkreader': {
        'github.com/jackc/pgproto3/v2',
        'github.com/jackc/pgproto3'
    },
    'github.com/creack/pty': {
        'github.com/kr/pty'
    },
    'github.com/rs/xid': {
        'github.com/rs/zerolog'
    },
    'github.com/zenazn/goji': {
        'github.com/rs/zerolog'
    },
    'github.com/jackc/pgproto3': {
        'github.com/jackc/pgx/v4'
    }
}

children = {
    'github.com/interuss/dss': {
        'github.com/golang-jwt/jwt',
        'github.com/golang-migrate/migrate/v4',
        'github.com/grpc-ecosystem/go-grpc-middleware',
        'google.golang.org/protobuf',
        'gopkg.in/square/go-jose.v2',
        'github.com/robfig/cron/v3',
        'github.com/golang/geo',
        'github.com/pkg/errors',
        'github.com/grpc-ecosystem/grpc-gateway',
        'github.com/stretchr/testify',
        'github.com/jonboulle/clockwork',
        'go.uber.org/zap',
        'github.com/lib/pq',
        'github.com/golang/protobuf',
        'github.com/interuss/stacktrace',
        'github.com/google/uuid',
        'github.com/coreos/go-semver',
        'github.com/cockroachdb/cockroach-go',
        'google.golang.org/genproto',
        'google.golang.org/grpc',
        'cloud.google.com/go'
    },
    'cloud.google.com/go': {
        'google.golang.org/protobuf',
        'github.com/google/pprof',
        'github.com/google/go-cmp',
        'golang.org/x/xerrors',
        'cloud.google.com/go/bigquery',
        'honnef.co/go/tools',
        'go.opencensus.io',
        'golang.org/x/oauth2',
        'golang.org/x/mod',
        'golang.org/x/sync',
        'golang.org/x/sys',
        'google.golang.org/api',
        'github.com/google/martian',
        'golang.org/x/time',
        'rsc.io/binaryregexp',
        'cloud.google.com/go/pubsub',
        'github.com/golang/groupcache',
        'golang.org/x/net',
        'golang.org/x/tools',
        'golang.org/x/text',
        'google.golang.org/appengine',
        'github.com/google/btree',
        'github.com/golang/protobuf',
        'golang.org/x/exp',
        'github.com/jstemmer/go-junit-report',
        'github.com/golang/mock',
        'cloud.google.com/go/storage',
        'google.golang.org/genproto',
        'golang.org/x/lint',
        'github.com/google/martian/v3',
        'github.com/googleapis/gax-go/v2',
        'google.golang.org/grpc',
        'cloud.google.com/go/datastore'
    },
    'github.com/golang-migrate/migrate/v4': {
        'modernc.org/b',
        'github.com/aws/aws-sdk-go',
        'github.com/fsouza/fake-gcs-server',
        'modernc.org/file',
        'github.com/ktrysmt/go-bitbucket',
        'github.com/golang/snappy',
        'modernc.org/zappy',
        'github.com/gobuffalo/here',
        'modernc.org/db',
        'github.com/bitly/go-hostpool',
        'golang.org/x/crypto',
        'modernc.org/fileutil',
        'github.com/bmizerany/assert',
        'modernc.org/sortutil',
        'modernc.org/strutil',
        'go.uber.org/atomic',
        'github.com/gorilla/mux',
        'github.com/tidwall/pretty',
        'github.com/xdg/stringprep',
        'golang.org/x/sys',
        'github.com/containerd/containerd',
        'github.com/stretchr/testify',
        'github.com/snowflakedb/gosnowflake',
        'google.golang.org/api',
        'github.com/neo4j/neo4j-go-driver',
        'github.com/edsrzf/mmap-go',
        'golang.org/x/time',
        'github.com/xanzy/go-gitlab',
        'modernc.org/lldb',
        'gitlab.com/nyarla/go-crypt',
        'modernc.org/internal',
        'github.com/markbates/pkger',
        'github.com/kardianos/osext',
        'modernc.org/ql',
        'github.com/jackc/pgconn',
        'github.com/cenkalti/backoff/v4',
        'github.com/xdg/scram',
        'github.com/lib/pq',
        'golang.org/x/net',
        'go.mongodb.org/mongo-driver',
        'golang.org/x/tools',
        'github.com/google/go-github',
        'github.com/sirupsen/logrus',
        'github.com/hashicorp/go-multierror',
        'modernc.org/golex',
        'github.com/mutecomm/go-sqlcipher/v4',
        'cloud.google.com/go/spanner',
        'github.com/gocql/gocql',
        'github.com/golang/protobuf',
        'github.com/denisenkom/go-mssqldb',
        'github.com/go-sql-driver/mysql',
        'github.com/ClickHouse/clickhouse-go',
        'github.com/cznic/mathutil',
        'github.com/dhui/dktest',
        'github.com/cockroachdb/cockroach-go',
        'cloud.google.com/go/storage',
        'google.golang.org/genproto',
        'github.com/mattn/go-sqlite3',
        'google.golang.org/grpc',
        'github.com/nakagami/firebirdsql',
        'github.com/docker/docker',
        'cloud.google.com/go'
    },
    'github.com/golang/protobuf': {
        'github.com/google/go-cmp',
        'google.golang.org/protobuf'
    },
    'github.com/grpc-ecosystem/go-grpc-middleware': {
        'github.com/gogo/protobuf',
        'github.com/go-logfmt/logfmt',
        'golang.org/x/oauth2',
        'github.com/go-kit/kit',
        'go.uber.org/atomic',
        'github.com/pkg/errors',
        'golang.org/x/sys',
        'github.com/go-stack/stack',
        'github.com/stretchr/testify',
        'go.uber.org/zap',
        'go.uber.org/multierr',
        'golang.org/x/net',
        'golang.org/x/text',
        'github.com/sirupsen/logrus',
        'github.com/golang/protobuf',
        'golang.org/x/exp',
        'google.golang.org/genproto',
        'google.golang.org/grpc',
        'github.com/opentracing/opentracing-go'
    },
    'github.com/grpc-ecosystem/grpc-gateway': {
        'golang.org/x/net',
        'github.com/antihax/optional',
        'github.com/golang/glog',
        'golang.org/x/xerrors',
        'google.golang.org/genproto',
        'golang.org/x/oauth2',
        'gopkg.in/yaml.v2',
        'github.com/golang/protobuf',
        'github.com/ghodss/yaml',
        'google.golang.org/grpc',
        'github.com/rogpeppe/fastuuid'
    },
    'github.com/stretchr/testify': {
        'github.com/pmezard/go-difflib',
        'github.com/stretchr/objx',
        'gopkg.in/yaml.v2',
        'gopkg.in/yaml.v3',
        'github.com/davecgh/go-spew'
    },
    'go.uber.org/zap': {
        'github.com/pkg/errors',
        'github.com/stretchr/testify',
        'honnef.co/go/tools',
        'go.uber.org/multierr',
        'golang.org/x/lint',
        'gopkg.in/yaml.v2',
        'go.uber.org/atomic'
    },
    'google.golang.org/genproto': {
        'golang.org/x/tools',
        'golang.org/x/net',
        'google.golang.org/protobuf',
        'honnef.co/go/tools',
        'golang.org/x/lint',
        'github.com/golang/protobuf',
        'golang.org/x/exp',
        'google.golang.org/grpc',
        'golang.org/x/sync'
    },
    'google.golang.org/grpc': {
        'google.golang.org/protobuf',
        'github.com/google/go-cmp',
        'github.com/golang/glog',
        'honnef.co/go/tools',
        'github.com/cncf/udpa/go',
        'golang.org/x/oauth2',
        'github.com/client9/misspell',
        'golang.org/x/sync',
        'golang.org/x/sys',
        'github.com/envoyproxy/go-control-plane',
        'golang.org/x/net',
        'golang.org/x/tools',
        'golang.org/x/text',
        'google.golang.org/appengine',
        'github.com/golang/protobuf',
        'github.com/google/uuid',
        'github.com/BurntSushi/toml',
        'github.com/envoyproxy/protoc-gen-validate',
        'github.com/golang/mock',
        'google.golang.org/genproto',
        'golang.org/x/lint',
        'cloud.google.com/go'
    },
    'google.golang.org/protobuf': {
        'github.com/google/go-cmp',
        'github.com/golang/protobuf',
        'google.golang.org/genproto'
    },
    'cloud.google.com/go/storage': {
        'github.com/google/go-cmp',
        'honnef.co/go/tools',
        'cloud.google.com/go/bigquery',
        'go.opencensus.io',
        'golang.org/x/oauth2',
        'golang.org/x/sys',
        'google.golang.org/api',
        'cloud.google.com/go/pubsub',
        'github.com/golang/groupcache',
        'golang.org/x/net',
        'golang.org/x/tools',
        'google.golang.org/appengine',
        'github.com/golang/protobuf',
        'golang.org/x/exp',
        'github.com/jstemmer/go-junit-report',
        'google.golang.org/genproto',
        'github.com/googleapis/gax-go/v2',
        'google.golang.org/grpc',
        'cloud.google.com/go/datastore',
        'cloud.google.com/go'
    },
    'github.com/golang/mock': {
        'golang.org/x/tools',
        'rsc.io/quote/v3'
    },
    'github.com/google/go-cmp': {
        'golang.org/x/xerrors'
    },
    'github.com/google/martian/v3': {
        'golang.org/x/net'
    },
    'github.com/google/pprof': {
        'github.com/ianlancetaylor/demangle',
        'golang.org/x/sys',
        'github.com/chzyer/test',
        'github.com/chzyer/readline',
        'github.com/chzyer/logex'
    },
    'github.com/googleapis/gax-go/v2': {
        'google.golang.org/grpc'
    },
    'go.opencensus.io': {
        'golang.org/x/net',
        'github.com/hashicorp/golang-lru',
        'golang.org/x/sys',
        'golang.org/x/text',
        'github.com/stretchr/testify',
        'github.com/google/go-cmp',
        'google.golang.org/genproto',
        'github.com/golang/protobuf',
        'google.golang.org/grpc',
        'github.com/golang/groupcache'
    },
    'golang.org/x/lint': {
        'golang.org/x/tools'
    },
    'golang.org/x/net': {
        'golang.org/x/sys',
        'golang.org/x/text',
        'golang.org/x/crypto'
    },
    'golang.org/x/oauth2': {
        'golang.org/x/net',
        'google.golang.org/appengine',
        'golang.org/x/sync',
        'cloud.google.com/go'
    },
    'golang.org/x/text': {
        'golang.org/x/tools'
    },
    'golang.org/x/tools': {
        'golang.org/x/net',
        'google.golang.org/appengine',
        'github.com/yuin/goldmark',
        'golang.org/x/xerrors',
        'golang.org/x/mod',
        'golang.org/x/sync'
    },
    'google.golang.org/api': {
        'golang.org/x/tools',
        'github.com/hashicorp/golang-lru',
        'golang.org/x/net',
        'golang.org/x/sys',
        'golang.org/x/text',
        'google.golang.org/appengine',
        'github.com/google/go-cmp',
        'honnef.co/go/tools',
        'google.golang.org/genproto',
        'go.opencensus.io',
        'golang.org/x/lint',
        'golang.org/x/oauth2',
        'github.com/googleapis/gax-go/v2',
        'github.com/golang/protobuf',
        'google.golang.org/grpc',
        'golang.org/x/sync',
        'cloud.google.com/go'
    },
    'cloud.google.com/go/spanner': {
        'golang.org/x/tools',
        'github.com/google/go-cmp',
        'google.golang.org/api',
        'golang.org/x/xerrors',
        'google.golang.org/genproto',
        'go.opencensus.io',
        'github.com/googleapis/gax-go/v2',
        'github.com/golang/protobuf',
        'google.golang.org/grpc',
        'cloud.google.com/go'
    },
    'github.com/ClickHouse/clickhouse-go': {
        'github.com/cloudflare/golz4',
        'github.com/stretchr/testify',
        'github.com/jmoiron/sqlx',
        'github.com/pierrec/lz4',
        'github.com/bkaradzic/go-lz4'
    },
    'github.com/aws/aws-sdk-go': {
        'github.com/jmespath/go-jmespath'
    },
    'github.com/denisenkom/go-mssqldb': {
        'golang.org/x/crypto',
        'github.com/golang-sql/civil'
    },
    'github.com/dhui/dktest': {
        'google.golang.org/protobuf',
        'github.com/Azure/go-ansiterm',
        'github.com/gogo/protobuf',
        'github.com/docker/go-units',
        'github.com/kr/pretty',
        'github.com/opencontainers/image-spec',
        'github.com/gorilla/mux',
        'github.com/pkg/errors',
        'golang.org/x/sys',
        'github.com/containerd/containerd',
        'github.com/stretchr/testify',
        'golang.org/x/time',
        'github.com/morikuni/aec',
        'gopkg.in/yaml.v2',
        'github.com/docker/go-connections',
        'golang.org/x/net',
        'github.com/docker/distribution',
        'github.com/lib/pq',
        'golang.org/x/text',
        'github.com/sirupsen/logrus',
        'github.com/gorilla/context',
        'github.com/golang/protobuf',
        'gotest.tools',
        'gopkg.in/check.v1',
        'github.com/opencontainers/go-digest',
        'google.golang.org/genproto',
        'github.com/Microsoft/go-winio',
        'google.golang.org/grpc',
        'github.com/docker/docker'
    },
    'github.com/fsouza/fake-gcs-server': {
        'github.com/gorilla/mux',
        'github.com/gorilla/handlers',
        'github.com/google/go-cmp',
        'google.golang.org/api',
        'github.com/sirupsen/logrus',
        'cloud.google.com/go/storage'
    },
    'github.com/gobuffalo/here': {
        'github.com/stretchr/testify',
        'gopkg.in/yaml.v2',
        'github.com/kr/pretty',
        'github.com/davecgh/go-spew',
        'gopkg.in/check.v1'
    },
    'github.com/gocql/gocql': {
        'github.com/hailocab/go-hostpool',
        'gopkg.in/inf.v0',
        'github.com/golang/snappy'
    },
    'github.com/hashicorp/go-multierror': {
        'github.com/hashicorp/errwrap'
    },
    'github.com/jackc/pgconn': {
        'github.com/pkg/errors',
        'golang.org/x/text',
        'github.com/stretchr/testify',
        'github.com/jackc/pgio',
        'github.com/jackc/pgmock',
        'golang.org/x/xerrors',
        'golang.org/x/crypto',
        'github.com/jackc/chunkreader/v2',
        'github.com/jackc/pgpassfile',
        'github.com/jackc/pgproto3/v2'
    },
    'github.com/ktrysmt/go-bitbucket': {
        'github.com/k0kubun/pp',
        'golang.org/x/net',
        'golang.org/x/sys',
        'google.golang.org/appengine',
        'github.com/mitchellh/mapstructure',
        'github.com/mattn/go-colorable',
        'github.com/mattn/go-isatty',
        'golang.org/x/oauth2',
        'github.com/golang/protobuf',
        'github.com/k0kubun/colorstring',
        'golang.org/x/sync'
    },
    'github.com/markbates/pkger': {
        'github.com/gobuffalo/here',
        'github.com/stretchr/testify'
    },
    'github.com/mutecomm/go-sqlcipher/v4': {
        'golang.org/x/net',
        'github.com/davecgh/go-spew',
        'github.com/stretchr/testify'
    },
    'github.com/neo4j/neo4j-go-driver': {
        'github.com/pkg/errors',
        'github.com/onsi/gomega',
        'github.com/golang/mock',
        'github.com/onsi/ginkgo'
    },
    'github.com/sirupsen/logrus': {
        'github.com/konsorten/go-windows-terminal-sequences',
        'github.com/pmezard/go-difflib',
        'golang.org/x/sys',
        'github.com/stretchr/testify',
        'github.com/stretchr/objx',
        'github.com/davecgh/go-spew'
    },
    'github.com/snowflakedb/gosnowflake': {
        'github.com/apache/arrow/go/arrow',
        'github.com/snowflakedb/glog',
        'github.com/dgrijalva/jwt-go',
        'github.com/pkg/browser',
        'golang.org/x/crypto',
        'github.com/google/uuid'
    },
    'github.com/xanzy/go-gitlab': {
        'golang.org/x/net',
        'github.com/google/go-querystring',
        'google.golang.org/appengine',
        'golang.org/x/oauth2',
        'golang.org/x/sync'
    },
    'golang.org/x/crypto': {
        'golang.org/x/net',
        'golang.org/x/sys'
    },
    'modernc.org/sortutil': {
        'modernc.org/mathutil',
        'github.com/remyoudompheng/bigfft'
    },
    'github.com/go-logfmt/logfmt': {
        'github.com/kr/logfmt'
    },
    'github.com/gogo/protobuf': {
        'github.com/kisielk/errcheck',
        'github.com/kisielk/gotool'
    },
    'golang.org/x/exp': {
        'golang.org/x/tools',
        'golang.org/x/sys',
        'github.com/go-gl/glfw',
        'golang.org/x/mobile',
        'golang.org/x/xerrors',
        'golang.org/x/mod',
        'dmitri.shuralyov.com/gpu/mtl',
        'golang.org/x/image',
        'github.com/BurntSushi/xgb',
        'github.com/go-gl/glfw/v3.3/glfw',
        'golang.org/x/sync'
    },
    'gopkg.in/yaml.v2': {
        'gopkg.in/check.v1'
    },
    'gopkg.in/yaml.v3': {
        'gopkg.in/check.v1'
    },
    'go.uber.org/atomic': {
        'golang.org/x/tools',
        'github.com/davecgh/go-spew',
        'golang.org/x/lint',
        'github.com/stretchr/testify'
    },
    'go.uber.org/multierr': {
        'golang.org/x/tools',
        'github.com/stretchr/testify',
        'honnef.co/go/tools',
        'golang.org/x/lint',
        'go.uber.org/tools',
        'go.uber.org/atomic'
    },
    'honnef.co/go/tools': {
        'golang.org/x/tools',
        'github.com/rogpeppe/go-internal',
        'golang.org/x/mod',
        'github.com/kisielk/gotool',
        'github.com/google/renameio',
        'github.com/BurntSushi/toml'
    },
    'github.com/cncf/udpa/go': {
        'github.com/golang/protobuf',
        'google.golang.org/grpc',
        'github.com/envoyproxy/protoc-gen-validate'
    },
    'github.com/envoyproxy/go-control-plane': {
        'google.golang.org/protobuf',
        'github.com/prometheus/client_model',
        'github.com/stretchr/testify',
        'github.com/google/go-cmp',
        'google.golang.org/genproto',
        'github.com/cncf/udpa/go',
        'github.com/golang/protobuf',
        'github.com/census-instrumentation/opencensus-proto',
        'google.golang.org/grpc',
        'github.com/envoyproxy/protoc-gen-validate'
    },
    'cloud.google.com/go/bigquery': {
        'golang.org/x/net',
        'golang.org/x/tools',
        'golang.org/x/sys',
        'google.golang.org/appengine',
        'github.com/google/go-cmp',
        'google.golang.org/api',
        'cloud.google.com/go/storage',
        'honnef.co/go/tools',
        'google.golang.org/genproto',
        'golang.org/x/lint',
        'github.com/golang/protobuf',
        'github.com/googleapis/gax-go/v2',
        'cloud.google.com/go/pubsub',
        'golang.org/x/exp',
        'google.golang.org/grpc',
        'golang.org/x/mod',
        'github.com/golang/groupcache',
        'cloud.google.com/go'
    },
    'google.golang.org/appengine': {
        'golang.org/x/net',
        'golang.org/x/tools',
        'golang.org/x/text',
        'golang.org/x/sys',
        'github.com/golang/protobuf',
        'golang.org/x/crypto'
    },
    'golang.org/x/mod': {
        'golang.org/x/tools',
        'golang.org/x/xerrors',
        'golang.org/x/crypto'
    },
    'github.com/jmoiron/sqlx': {
        'github.com/mattn/go-sqlite3',
        'github.com/lib/pq',
        'github.com/go-sql-driver/mysql'
    },
    'github.com/Microsoft/go-winio': {
        'github.com/pkg/errors',
        'golang.org/x/sys',
        'github.com/sirupsen/logrus'
    },
    'github.com/kr/pretty': {
        'github.com/kr/text'
    },
    'github.com/jackc/pgmock': {
        'github.com/jackc/pgconn',
        'github.com/stretchr/testify',
        'golang.org/x/xerrors',
        'github.com/jackc/pgtype',
        'github.com/kr/pretty',
        'github.com/jackc/pgproto3/v2',
        'gopkg.in/check.v1'
    },
    'github.com/jackc/pgpassfile': {
        'github.com/stretchr/testify'
    },
    'github.com/jackc/pgproto3/v2': {
        'github.com/pkg/errors',
        'github.com/stretchr/testify',
        'github.com/jackc/pgio',
        'github.com/jackc/chunkreader/v2',
        'github.com/jackc/chunkreader'
    },
    'github.com/onsi/ginkgo': {
        'github.com/hpcloud/tail',
        'github.com/onsi/gomega',
        'golang.org/x/sys'
    },
    'github.com/onsi/gomega': {
        'gopkg.in/fsnotify.v1',
        'golang.org/x/net',
        'gopkg.in/tomb.v1',
        'golang.org/x/sys',
        'golang.org/x/text',
        'github.com/hpcloud/tail',
        'github.com/onsi/ginkgo',
        'golang.org/x/xerrors',
        'gopkg.in/yaml.v2',
        'github.com/golang/protobuf',
        'github.com/fsnotify/fsnotify',
        'golang.org/x/sync'
    },
    'github.com/apache/arrow/go/arrow': {
        'github.com/pmezard/go-difflib',
        'github.com/stretchr/testify',
        'golang.org/x/xerrors',
        'github.com/google/flatbuffers',
        'github.com/davecgh/go-spew'
    },
    'github.com/kisielk/errcheck': {
        'golang.org/x/tools',
        'github.com/kisielk/gotool'
    },
    'golang.org/x/image': {
        'golang.org/x/text'
    },
    'golang.org/x/mobile': {
        'golang.org/x/exp',
        'golang.org/x/image',
        'golang.org/x/sys'
    },
    'github.com/rogpeppe/go-internal': {
        'gopkg.in/errgo.v2'
    },
    'github.com/prometheus/client_model': {
        'github.com/golang/protobuf',
        'golang.org/x/sync'
    },
    'cloud.google.com/go/datastore': {
        'golang.org/x/tools',
        'golang.org/x/sys',
        'google.golang.org/grpc',
        'google.golang.org/appengine',
        'github.com/google/go-cmp',
        'google.golang.org/api',
        'google.golang.org/genproto',
        'github.com/googleapis/gax-go/v2',
        'github.com/golang/protobuf',
        'cloud.google.com/go/pubsub',
        'golang.org/x/exp',
        'cloud.google.com/go'
    },
    'cloud.google.com/go/pubsub': {
        'github.com/google/go-cmp',
        'cloud.google.com/go/bigquery',
        'go.opencensus.io',
        'golang.org/x/oauth2',
        'golang.org/x/mod',
        'golang.org/x/sync',
        'golang.org/x/sys',
        'google.golang.org/api',
        'golang.org/x/time',
        'golang.org/x/net',
        'golang.org/x/tools',
        'github.com/golang/protobuf',
        'golang.org/x/exp',
        'cloud.google.com/go/storage',
        'google.golang.org/genproto',
        'golang.org/x/lint',
        'github.com/googleapis/gax-go/v2',
        'google.golang.org/grpc',
        'cloud.google.com/go'
    },
    'github.com/kr/text': {
        'github.com/kr/pty'
    },
    'github.com/jackc/pgtype': {
        'github.com/lib/pq',
        'github.com/stretchr/testify',
        'github.com/jackc/pgio',
        'github.com/jackc/pgx/v4',
        'golang.org/x/xerrors',
        'github.com/shopspring/decimal',
        'go.uber.org/multierr',
        'github.com/kr/pretty',
        'github.com/satori/go.uuid',
        'gopkg.in/check.v1'
    },
    'rsc.io/quote/v3': {
        'rsc.io/sampler'
    },
    'gopkg.in/errgo.v2': {
        'gopkg.in/check.v1',
        'github.com/kr/pretty'
    },
    'github.com/jackc/pgx/v4': {
        'github.com/jackc/pgio',
        'golang.org/x/xerrors',
        'github.com/jackc/pgproto3',
        'golang.org/x/crypto',
        'github.com/kr/pretty',
        'github.com/mattn/go-colorable',
        'gopkg.in/inconshreveable/log15.v2',
        'go.uber.org/atomic',
        'github.com/pkg/errors',
        'github.com/kr/pty',
        'github.com/konsorten/go-windows-terminal-sequences',
        'golang.org/x/sys',
        'github.com/go-stack/stack',
        'github.com/stretchr/testify',
        'go.uber.org/zap',
        'go.uber.org/multierr',
        'github.com/satori/go.uuid',
        'github.com/cockroachdb/apd',
        'github.com/jackc/pgconn',
        'golang.org/x/net',
        'golang.org/x/tools',
        'github.com/lib/pq',
        'golang.org/x/text',
        'github.com/sirupsen/logrus',
        'github.com/shopspring/decimal',
        'github.com/mattn/go-isatty',
        'github.com/jackc/puddle',
        'gopkg.in/check.v1',
        'github.com/coreos/go-systemd',
        'github.com/stretchr/objx',
        'github.com/jackc/pgtype',
        'github.com/rs/zerolog',
        'github.com/jackc/pgproto3/v2'
    },
    'rsc.io/sampler': {
        'golang.org/x/text'
    },
    'github.com/jackc/puddle': {
        'github.com/stretchr/testify'
    },
    'github.com/kr/pty': {
        'github.com/creack/pty'
    },
    'github.com/mattn/go-colorable': {
        'github.com/mattn/go-isatty'
    },
    'github.com/mattn/go-isatty': {
        'golang.org/x/sys'
    },
    'github.com/rs/zerolog': {
        'github.com/pkg/errors',
        'golang.org/x/tools',
        'github.com/coreos/go-systemd',
        'github.com/rs/xid',
        'github.com/zenazn/goji'
    },
    'github.com/stretchr/objx': {
        'github.com/davecgh/go-spew',
        'github.com/stretchr/testify'
    },
    'github.com/jackc/pgproto3': {
        'github.com/pkg/errors',
        'github.com/jackc/pgio',
        'github.com/jackc/chunkreader'
    }
}

def find_all_paths_parent_to_target(parent, child, curr_path, all_paths):
    if child not in parents:
        print('destination not found')
    for p in parents[child]:
        if p not in curr_path:
            if p == parent:
                all_paths.append(curr_path+[p])
            else:
                find_all_paths_parent_to_target(parent, p, curr_path+[p], all_paths)



def get_path(parents, dest, base, path, all_paths):
    if dest not in path:
        path.append(dest)
    if dest in parents:
        for parent in parents[dest]:
            curr_path = copy.copy(path)
            curr_path.append(parent)
            if parent == base:
                all_paths.append(curr_path)
            else:
                get_path(parents, parent, base, curr_path, all_paths)

def get_all_paths():
    base = 'github.com/interuss/dss'
    dest = 'github.com/containerd/containerd'
    all_paths = []
    get_path(parents, dest, base, [], all_paths)
    print('all_paths: ', all_paths)
    # return all_paths

def main():
    get_all_paths()
            

if __name__ == '__main__':
    main()