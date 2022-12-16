package internal

import (
	"embed"
	"net/http"
	"time"

	"github.com/d2jvkpn/crons/crons"

	"github.com/d2jvkpn/go-web/pkg/wrap"
)

const (
	HTTP_MaxHeaderBytes     = 2 << 11 // 4K
	HTTP_ReadHeaderTimeout  = 2 * time.Second
	HTTP_ReadTimeout        = 10 * time.Second
	HTTP_WriteTimeout       = 10 * time.Second
	HTTP_IdleTimeout        = 60
	HTTP_MaxMultipartMemory = 8 << 20 // 8M
)

var (
	//go:embed static
	_Static embed.FS
	//go:embed templates
	_Templates embed.FS

	_Release bool
	_Server  *http.Server
	Logger   *wrap.Logger
	Manager  *crons.Manager
)
