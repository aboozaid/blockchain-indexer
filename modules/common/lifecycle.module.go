package common

type LifecycleModule interface {
	OnAppStart() error
	OnAppDestroy() error
}

type LifecycleService interface {
	OnModuleStart() error
	OnModuleStop() error
}