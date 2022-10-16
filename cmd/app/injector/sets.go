package injector

import (
	"user-manager/repository"
	"user-manager/service"

	"github.com/google/wire"
)

var EmailTemplates = wire.NewSet(ProvideBaseTemplate, ProvideTranslations)
var RequestContext = wire.NewSet(ProvideTx, ProvideLog, ProvideSecurityLog, ProvideUserSession)
var Repositories = wire.NewSet(repository.ProvideMailQueueRepository, repository.ProvideSessionRepository, repository.ProvideUserRepository, repository.ProvideSecondFactorThrottlingRepository)
var Services = wire.NewSet(service.ProvideMailQueueService, service.ProvideSessionCookieService, service.ProvideAuthService)

var AllDependencies = wire.NewSet(ProvideDatabase, ProvideConfig, EmailTemplates, RequestContext, Repositories, Services)
