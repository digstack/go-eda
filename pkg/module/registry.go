package module

import (
	"context"
	"fmt"

	"github.com/yourusername/go-generic-event-driven/pkg/cqrs"
	"github.com/yourusername/go-generic-event-driven/pkg/ddd"
	"github.com/yourusername/go-generic-event-driven/pkg/logger"
)

// Module represents a generic module with handlers
type Module struct {
	Name string

	CommandHandlers map[string]cqrs.CommandHandler
	QueryHandlers   map[string]cqrs.QueryHandler
	EventHandlers   map[string][]cqrs.EventHandler

	Repositories map[string]interface{}
	Services     map[string]interface{}

	Aggregates []ddd.Aggregate
	Events     []ddd.EventWithMetadata
}

func NewModule(name string) *Module {
	return &Module{
		Name: name,

		CommandHandlers: make(map[string]cqrs.CommandHandler),
		QueryHandlers:   make(map[string]cqrs.QueryHandler),
		EventHandlers:   make(map[string][]cqrs.EventHandler),

		Repositories: make(map[string]interface{}),
		Services:     make(map[string]interface{}),

		Aggregates: make([]ddd.Aggregate, 0),
		Events:     make([]ddd.EventWithMetadata, 0),
	}
}

func (m *Module) RegisterCommandHandler(commandType string, handler cqrs.CommandHandler) error {
	if _, exists := m.CommandHandlers[commandType]; exists {
		return fmt.Errorf("command handler %s already registered in module %s", commandType, m.Name)
	}
	m.CommandHandlers[commandType] = handler
	return nil
}

func (m *Module) RegisterQueryHandler(queryType string, handler cqrs.QueryHandler) error {
	if _, exists := m.QueryHandlers[queryType]; exists {
		return fmt.Errorf("query handler %s already registered in module %s", queryType, m.Name)
	}
	m.QueryHandlers[queryType] = handler
	return nil
}

func (m *Module) RegisterEventHandler(eventType string, handler cqrs.EventHandler) error {
	if _, exists := m.EventHandlers[eventType]; !exists {
		m.EventHandlers[eventType] = make([]cqrs.EventHandler, 0)
	}
	m.EventHandlers[eventType] = append(m.EventHandlers[eventType], handler)
	return nil
}

func (m *Module) RegisterRepository(name string, repo interface{}) error {
	if _, exists := m.Repositories[name]; exists {
		return fmt.Errorf("repository %s already registered in module %s", name, m.Name)
	}
	m.Repositories[name] = repo
	return nil
}

func (m *Module) RegisterService(name string, service interface{}) error {
	if _, exists := m.Services[name]; exists {
		return fmt.Errorf("service %s already registered in module %s", name, m.Name)
	}
	m.Services[name] = service
	return nil
}

func (m *Module) RegisterAggregate(aggregate ddd.Aggregate) {
	m.Aggregates = append(m.Aggregates, aggregate)
}

func (m *Module) RegisterEvent(event ddd.EventWithMetadata) {
	m.Events = append(m.Events, event)
}

// Registry manages multiple modules
type Registry struct {
	modules map[string]*Module
	logger  logger.Logger
}

func NewRegistry(logger logger.Logger) *Registry {
	return &Registry{
		modules: make(map[string]*Module),
		logger:  logger,
	}
}

func (r *Registry) Register(module *Module) error {
	if _, exists := r.modules[module.Name]; exists {
		return fmt.Errorf("module %s already registered", module.Name)
	}

	r.modules[module.Name] = module
	r.logger.Info("Module registered", logger.NewField("module", module.Name))
	return nil
}

func (r *Registry) GetModule(name string) (*Module, bool) {
	module, exists := r.modules[name]
	return module, exists
}

func (r *Registry) GetAllModules() map[string]*Module {
	return r.modules
}

func (r *Registry) SetupBuses(
	commandBus cqrs.CommandBus,
	queryBus cqrs.QueryBus,
	eventBus cqrs.EventBus,
) error {
	for _, module := range r.modules {
		if err := r.setupModuleBuses(module, commandBus, queryBus, eventBus); err != nil {
			return fmt.Errorf("failed to setup buses for module %s: %w", module.Name, err)
		}
	}
	return nil
}

func (r *Registry) setupModuleBuses(
	module *Module,
	commandBus cqrs.CommandBus,
	queryBus cqrs.QueryBus,
	eventBus cqrs.EventBus,
) error {
	// Register command handlers
	for commandType, handler := range module.CommandHandlers {
		if err := commandBus.Register(commandType, handler); err != nil {
			return fmt.Errorf("failed to register command handler %s: %w", commandType, err)
		}
		r.logger.Debug("Command handler registered",
			logger.String("module", module.Name),
			logger.String("command", commandType))
	}

	// Register query handlers
	for queryType, handler := range module.QueryHandlers {
		if err := queryBus.Register(queryType, handler); err != nil {
			return fmt.Errorf("failed to register query handler %s: %w", queryType, err)
		}
		r.logger.Debug("Query handler registered",
			logger.String("module", module.Name),
			logger.String("query", queryType))
	}

	// Register event handlers
	for eventType, handlers := range module.EventHandlers {
		for _, handler := range handlers {
			if err := eventBus.Subscribe(eventType, handler); err != nil {
				return fmt.Errorf("failed to register event handler %s: %w", eventType, err)
			}
			r.logger.Debug("Event handler registered",
				logger.String("module", module.Name),
				logger.String("event", eventType))
		}
	}

	r.logger.Info("Module buses setup completed", logger.String("module", module.Name))
	return nil
}

// ModuleInitializer defines how to initialize a module
type ModuleInitializer interface {
	Initialize(ctx context.Context, module *Module) error
}

// Application represents the entire application with modules
type Application struct {
	Registry *Registry
	CQRS     *cqrs.CQRS
	Logger   logger.Logger
}

func NewApplication(logger logger.Logger) *Application {
	commandBus := cqrs.NewInMemoryCommandBus()
	queryBus := cqrs.NewInMemoryQueryBus()
	eventBus := cqrs.NewInMemoryEventBus()

	return &Application{
		Registry: NewRegistry(logger),
		CQRS:     cqrs.NewCQRS(commandBus, queryBus, eventBus),
		Logger:   logger,
	}
}

func (app *Application) RegisterModule(module *Module) error {
	return app.Registry.Register(module)
}

func (app *Application) Start(ctx context.Context) error {
	app.Logger.Info("Starting application")

	// Setup all module buses
	if err := app.Registry.SetupBuses(
		app.CQRS.CommandBus,
		app.CQRS.QueryBus,
		app.CQRS.EventBus,
	); err != nil {
		return fmt.Errorf("failed to setup module buses: %w", err)
	}

	app.Logger.Info("Application started successfully")
	return nil
}

func (app *Application) Stop() error {
	app.Logger.Info("Stopping application")
	// Cleanup logic here
	app.Logger.Info("Application stopped")
	return nil
}
