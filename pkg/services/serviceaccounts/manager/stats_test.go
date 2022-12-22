package manager

import (
	"context"
	"testing"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/serviceaccounts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UsageStats(t *testing.T) {
	storeMock := newServiceAccountStoreFake()
	svc := ServiceAccountsService{storeMock, log.New("test"), log.New("background-test"), &SecretsCheckerFake{}, true, 5}
	err := svc.DeleteServiceAccount(context.Background(), 1, 1)
	require.NoError(t, err)

	storeMock.ExpectedStats = &serviceaccounts.Stats{
		ServiceAccounts: 1,
		Tokens:          1,
	}
	stats, err := svc.getUsageMetrics(context.Background())
	require.NoError(t, err)

	assert.Equal(t, int64(1), stats["stats.serviceaccounts.count"].(int64))
	assert.Equal(t, int64(1), stats["stats.serviceaccounts.tokens.count"].(int64))
	assert.Equal(t, int64(1), stats["stats.serviceaccounts.secret_scan.enabled.count"].(int64))
}
