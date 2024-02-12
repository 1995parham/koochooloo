package urlsvc_test

import (
	"context"
	"testing"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
	"github.com/1995parham/koochooloo/internal/domain/service/urlsvc"
	"github.com/1995parham/koochooloo/internal/infra/config"
	"github.com/1995parham/koochooloo/internal/infra/generator"
	"github.com/1995parham/koochooloo/internal/infra/logger"
	"github.com/1995parham/koochooloo/internal/infra/repository/urldb"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type URLSuite struct {
	suite.Suite

	svc *urlsvc.URLSvc
	app *fxtest.App
}

func (suite *URLSuite) SetupSuite() {
	suite.app = fxtest.New(suite.T(),
		fx.Provide(config.Provide),
		fx.Provide(logger.Provide),
		fx.Provide(telemetry.ProvideNull),
		fx.Provide(generator.Provide),
		fx.Provide(
			fx.Annotate(urldb.ProvideMemory, fx.As(new(urlrepo.Repository))),
		),
		fx.Provide(urlsvc.Provide),
		fx.Invoke(func(svc *urlsvc.URLSvc) {
			suite.svc = svc
		}),
	).RequireStart()
}

func (suite *URLSuite) TearDownSuite() {
	suite.app.RequireStop()
}

// nolint: funlen
func (suite *URLSuite) TestSetGetCount() {
	require := suite.Require()

	cases := []struct {
		name           string
		key            string
		url            string
		expire         time.Time
		expectedSetErr error
		expectedGetErr error
	}{
		{
			name:           "Successful",
			key:            "raha",
			url:            "https://elahe-dastan.github.io",
			expire:         time.Time{},
			expectedSetErr: nil,
			expectedGetErr: nil,
		},
		{
			name:           "Duplicate Key",
			key:            "raha",
			url:            "https://elahe-dastan.github.io",
			expire:         time.Time{},
			expectedSetErr: urlrepo.ErrDuplicateKey,
			expectedGetErr: nil,
		},
		{
			name:           "Automatic",
			key:            "",
			url:            "https://1995parham.me",
			expire:         time.Time{},
			expectedSetErr: nil,
			expectedGetErr: nil,
		},
		{
			name:           "Expired",
			key:            "",
			url:            "https://github.com",
			expire:         time.Now().Add(-time.Minute),
			expectedSetErr: nil,
			expectedGetErr: urlrepo.ErrKeyNotFound,
		},
	}

	for _, c := range cases {
		c := c
		suite.Run(c.name, func() {
			expire := &c.expire
			if c.expire.IsZero() {
				expire = nil
			}

			key, err := suite.svc.Set(context.Background(), c.key, c.url, expire, 0)
			require.ErrorIs(
				err,
				c.expectedSetErr,
			)

			if c.expectedSetErr == nil {
				if c.key != "" {
					require.Equal("$"+c.key, key)
				} else {
					require.Len(key, 6)
				}

				url, err := suite.svc.Get(context.Background(), key)
				require.ErrorIs(err, c.expectedGetErr)

				if c.expectedGetErr == nil {
					require.Equal(c.url, url)
				}

				count, err := suite.svc.Count(context.Background(), key)
				require.ErrorIs(err, c.expectedGetErr)

				if c.expectedGetErr == nil {
					require.Equal(0, count)
				}
			}
		})
	}
}

func TestURLSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(URLSuite))
}
