package buffalo

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/envy"
	"github.com/gorilla/sessions"
	"github.com/markbates/going/defaults"
	"github.com/markbates/pop"
)

// Options are used to configure and define how your application should run.
type Options struct {
	Name string
	// Env is the "environment" in which the App is running. Default is "development".
	Env string
	// LogLevel defaults to "debug".
	LogLevel string
	// Logger to be used with the application. A default one is provided.
	Logger Logger
	// MethodOverride allows for changing of the request method type. See the default
	// implementation at buffalo.MethodOverride
	MethodOverride http.HandlerFunc
	// SessionStore is the `github.com/gorilla/sessions` store used to back
	// the session. It defaults to use a cookie store and the ENV variable
	// `SESSION_SECRET`.
	SessionStore sessions.Store
	// SessionName is the name of the session cookie that is set. This defaults
	// to "_buffalo_session".
	SessionName string
	// Host that this application will be available at. Default is "http://127.0.0.1:[$PORT|3000]".
	Host string
	// Worker implements the Worker interface and can process tasks in the background.
	// Default is "github.com/gobuffalo/worker.Simple.
	Worker worker.Worker
	// WorkerOff tells App.Start() whether to start the Worker process or not. Default is "false".
	WorkerOff bool

	// PreHandlers are http.Handlers that are called between the http.Server
	// and the buffalo Application.
	PreHandlers []http.Handler
	// PreWare takes an http.Handler and returns and http.Handler
	// and acts as a pseudo-middleware between the http.Server and
	// a Buffalo application.
	PreWares []PreWare

	Context context.Context
	cancel  context.CancelFunc
	Prefix  string
}

// PreWare takes an http.Handler and returns and http.Handler
// and acts as a pseudo-middleware between the http.Server and
// a Buffalo application.
type PreWare func(http.Handler) http.Handler

// NewOptions returns a new Options instance with sensible defaults
func NewOptions() Options {
	return optionsWithDefaults(Options{})
}

func optionsWithDefaults(opts Options) Options {
	opts.Env = defaults.String(opts.Env, envy.Get("GO_ENV", "development"))
	opts.LogLevel = defaults.String(opts.LogLevel, "debug")
	opts.Name = defaults.String(opts.Name, "/")

	if opts.PreWares == nil {
		opts.PreWares = []PreWare{}
	}
	if opts.PreHandlers == nil {
		opts.PreHandlers = []http.Handler{}
	}

	if opts.Context == nil {
		opts.Context = context.Background()
	}
	opts.Context, opts.cancel = context.WithCancel(opts.Context)

	if opts.Logger == nil {
		opts.Logger = NewLogger(opts.LogLevel)
	}

	pop.Log = func(s string, args ...interface{}) {
		if pop.Debug {
			l := opts.Logger
			if len(args) > 0 {
				for i, a := range args {
					l = l.WithField(fmt.Sprintf("$%d", i+1), a)
				}
			}
			if pop.Color {
				s = color.YellowString(s)
			}
			l.Debug(s)
		}
	}

	if opts.SessionStore == nil {
		secret := envy.Get("SESSION_SECRET", "")
		// In production a SESSION_SECRET must be set!
		if opts.Env == "production" && secret == "" {
			log.Println("WARNING! Unless you set SESSION_SECRET env variable, your session storage is not protected!")
		}
		opts.SessionStore = sessions.NewCookieStore([]byte(secret))
	}
	if opts.Worker == nil {
		w := worker.NewSimpleWithContext(opts.Context)
		w.Logger = opts.Logger
		opts.Worker = w
	}
	opts.SessionName = defaults.String(opts.SessionName, "_buffalo_session")
	opts.Host = defaults.String(opts.Host, envy.Get("HOST", fmt.Sprintf("http://127.0.0.1:%s", envy.Get("PORT", "3000"))))
	return opts
}
