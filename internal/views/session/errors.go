package session

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services"
)

// Warn shows an error as a warning notice, in words the user can act on.
func (session *Session) Warn(err error) { session.Notify(Describe(err), true) }

// Describe turns the errors the user can cause into Spanish sentences; anything else keeps its own text.
func Describe(err error) string {
	switch {
	case errors.Is(err, domain.ErrUnsupportedScript):
		return "No sé cómo ejecutar ese archivo: usa .sh, .py, .js o .ts, o dale permiso de ejecución."
	case errors.Is(err, services.ErrAlreadyRunning):
		return "Ya está en marcha."
	case errors.Is(err, services.ErrNotAFile):
		return "Eso no es un archivo."
	case errors.Is(err, services.ErrNoPython):
		return "No encuentro python3 y no puedo descargarlo para este ordenador. Instálalo y vuelve a intentarlo."
	case errors.Is(err, domain.ErrEmptyAlias):
		return "Escribe un nombre."
	case errors.Is(err, domain.ErrInvalidAlias):
		return fmt.Sprintf("Usa solo letras, números, - y _, sin espacios y hasta %d caracteres.", domain.MaxAliasLength)
	case errors.Is(err, domain.ErrAliasTaken):
		return "Ya existe uno con ese nombre: elige otro."
	case errors.Is(err, services.ErrTemplateExists):
		return "Esa plantilla ya está en la carpeta del proyecto: borra el archivo si quieres guardarla otra vez."
	case errors.Is(err, fs.ErrNotExist):
		return "El archivo o la carpeta ya no existe."
	default:
		return err.Error()
	}
}
