package handler

import (
	"net/http"
	"net/url"

	"github.com/zhirsch/destinykioskstatus/src/db"
	"github.com/zhirsch/destinykioskstatus/src/kiosk"
	"github.com/zhirsch/destinykioskstatus/src/server"
)

type VendorHandler struct {
	Server     *server.Server
	VendorHash uint32
}

type Data struct {
	kiosk.Data
	Characters       []Character
	CurrentCharacter string
}

type Character struct {
	ID      string
	Class   string
	Current bool
	URL     string
}

func (h VendorHandler) ServeHTTP(bungieUser *db.BungieUser, w http.ResponseWriter, r *http.Request) {
	// TODO: Support multiple DestinyUsers on the same BungieUser.
	destinyUser := bungieUser.DestinyUsers[0]

	// Get the character to display info for.  If there isn't a character,
	// redirect to the first character.
	characterID := db.DestinyCharacterID(r.URL.Query().Get("c"))
	if characterID == "" {
		http.Redirect(w, r, characterURL(*r.URL, destinyUser.DestinyCharacters[0]), http.StatusFound)
		return
	}

	data := Data{
		Data:             kiosk.FetchKioskStatus(bungieUser, destinyUser, characterID, h.VendorHash, h.Server.API, h.Server.Manifest),
		CurrentCharacter: string(characterID),
	}
	for _, character := range destinyUser.DestinyCharacters {
		data.Characters = append(data.Characters, Character{
			ID:      string(character.CharacterID),
			Class:   character.ClassName,
			Current: character.CharacterID == characterID,
			URL:     characterURL(*r.URL, character),
		})
	}

	if err := h.Server.Template.Execute(w, data); err != nil {
		panic(err)
	}
}

func characterURL(u url.URL, destinyCharacter *db.DestinyCharacter) string {
	q := u.Query()
	q.Set("c", string(destinyCharacter.CharacterID))
	u.RawQuery = q.Encode()
	return u.String()
}
