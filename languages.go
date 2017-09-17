/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2017 Mnpn

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
// TRANSLATORS:
// - Swedish, Mnpn03
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/legolord208/stdutil"
)

var errLangCorrupt = errors.New("corrupt language file")
var lang map[string]string

// TL stands for TransLate kek
func tl(name string) string {
	str, ok := lang[name]
	if ok {
		return str
	}

	return name
}

func loadLangAuto(langfile string) {
	fmt.Println("Loading language...")
	switch langfile {
	case "en":
		loadLangDefault()
	case "sv":
		loadLangString(langSv)
	default:
		reader, err := os.Open(langfile)
		if err != nil {
			stdutil.PrintErr("Could not read language file", err)
			return
		}
		defer reader.Close()

		err = loadLang(reader)
		if err != nil {
			stdutil.PrintErr("Could not load language file", err)
			loadLangDefault()
		}
	}
}
func loadLang(reader io.Reader) error {
	lang = make(map[string]string)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		parts := strings.SplitN(text, "=", 2)
		if len(parts) != 2 {
			return errLangCorrupt
		}
		key := parts[0]
		val := parts[1]

		if strings.HasSuffix(key, ".dev") && devVersion {
			key = key[:len(key)-len(".dev")]
		}

		lang[key] = val
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
func loadLangString(lang string) error {
	return loadLang(strings.NewReader(lang))
}
func loadLangDefault() {
	loadLangString(langEn)
}

// Here is just some long data.
// This comment is a separator, btw.

var langEn = `
update.checking=Checking for updates...
update.error=Error checking for updates
update.available=Update available! Version
update.available.dev=Latest stable release:
update.download=Download from
update.none=No updates found.

loading.bookmarks=Reading bookmarks...

failed.generic=Failed
failed.reading=Could not read
failed.readline.start=Could not start readline library
failed.readline.read=Could not read line
failed.auth=Couldn't authenticate
failed.session.start=Could not open session
failed.perms=No permissions to perform this action.
failed.path.home=Could not determine value of ~
failed.user=Couldn't query user
failed.user.edit=Couldn't edit user data
failed.channel=Could not query channel
failed.guild=Could not query guild
failed.guild.edit=Could not edit guild
failed.settings=Could not query user settings
failed.timestamp=Couldn't parse timestamp
failed.channel.create=Could not create channel
failed.msg.query=Could not get message
failed.msg.send=Could not send message
failed.msg.edit=Couldn't edit message
failed.msg.delete=Couldn't delete message
failed.lua.run=Could not run lua
failed.lua.event=Recovered from LUA error
failed.voice.connect=Could not connect to voice channel
failed.voice.speak=Could not start speaking
failed.voice.disconnect=Could not disconnect
failed.voice.regions=Could not get regions
failed.exec=Could not execute
failed.fixpath=Could not 'fix' filepath
failed.file.open=Couldn't open file
failed.file.write=Could not write file
failed.file.read=Could not read file
failed.file.load=Could not load file
failed.file.save=Could not save file
failed.file.delete=Could not delete file
failed.status=Couldn't update status
failed.typing=Couldn't start typing
failed.members=Could not list members
failed.invite=Could not query invite
failed.invite.accept=Could not accept invite
failed.invite.create=Invite could not be created
failed.roles=Could not get roles
failed.role.change=Could not add/remove role
failed.role.create=Could not create role
failed.role.edit=Could not edit role
failed.role.delete=Could not delete role!
failed.nick=Could not set nickname
failed.ban.create=Could not ban user
failed.ban.delete=Could not unban user
failed.ban.list=Could not list bans
failed.kick=Could not kick user
failed.leave=Could not leave
failed.block=Couldn't block user
failed.friends=Couldn't get friends
failed.json=Could not parse json
failed.base64=Couldn't convert to Base64
failed.react=Could not react to message
failed.react.used=Emoji used already, skipping
failed.webrequest=Could not make web request
failed.avatar=Couldn't set avatar
failed.status=Could not set status
failed.api.start=Couldn't start API
failed.mfa=The account has 2FA enabled. Consider using a user token instead
failed.permcalc=Could not open permission calculator
failed.nochannel=Server does not have a channel

invalid.yn=Please type either 'y' or 'n'.
invalid.webhook=Webhook format invalid. Format: id/token
invalid.webhook.command=Not an allowed webhook command
invalid.limit.message=Message exceeds character limit
invalid.channel=No channel selected!
invalid.channel.voice=No voice channel selected!
invalid.guild=No guild selected!
invalid.id=You need to select something to get its ID!
invalid.value=No such value
invalid.role=No role with that ID
invalid.number=Not a number
invalid.cache=No cache available!
invalid.onlyfor.users=This only works for users.
invalid.onlyfor.bots=This command only works for bot users.
invalid.music.playing=Already playing something
invalid.bookmark=Bookmark doesn't exist
invalid.status.offline=The offline status exists, but cannot be set through the API
invalid.command=Unknown command:
invalid.command2=Do 'help' for help
invalid.api.started=API already started
invalid.api.notstarted=API not started
invalid.source.terminal=You must be in terminal to do this.

login.token=Please paste your 'token' here.
login.token.user=User tokens are prefixed with 'user '
login.token.webhook=Webhook tokens are prefixed with 'webhook ', and their URL or id/token
login.starting=Authenticating...
login.hidden=[Hidden]
login.finish=Logged in with user ID
intro.help=Write 'help' for help
intro.exit=Press Ctrl+D or type 'exit' to exit.

pointer.unknown=Unknown
pointer.private=Private

status.msg.create=Created message with ID
status.msg.delall=Deleted # messages
status.channel=Selected channel with ID
status.invite.accept=Accepted invite.
status.invite.create=Created invite with code:
status.cache=Message cached!
status.loading=Loading...
status.avatar=Avatar set!
status.name=Name set!
status.status=Status set!
status.api.start=API started:

rl.session=Restarting session...
rl.cache.loc=Reloading location cache...
rl.cache.vars=Deleting cache variables...
`

var langSv = `
update.checking=Letar efter uppdateringar...
update.error=Ett fel inträffade under letandet efter nya uppdateringar.
update.available=En uppdatering finns tillgänglig! Version
update.available.dev=Senaste stabila version:
update.download=Ladda ner från
update.none=Hittade inga uppdateringar.
loading.bookmarks=Läser bokmärken...
failed.reading=Kunde inte läsa
failed.readline.start=Kunde inte starta "readline"-biblioteket
failed.readline.read=Kunde inte läsa linje
failed.auth=Kunde inte autentisera
failed.session.start=Kunde inte öppna sessionen
failed.perms=Inga behörigheter för att utföra den här åtgärden.
failed.path.home=Det gick inte att bedöma värdet av ~
failed.user=Kunde inte förfråga efter användarinformation
failed.user.edit=Kunde inte ändra användarinformationen
failed.channel=Kunde inte fråga efter kanal
failed.guild=Kunde inte fråga efter server
failed.timestamp=Kunde inte tolka tidsstämplar
failed.channel.create=Kunde inte skapa kanal
failed.msg.query=Kunde inte ta emot meddelande
failed.msg.send=Kunde inte skicka meddelande
failed.msg.edit=Kunde inte ändra meddelande
failed.msg.delete=Kunde inte ta bort meddelande
failed.lua.run=Kunde inte köra lua
failed.lua.event=Återhämtade från LUA-fel
failed.voice.connect=Kunde inte ansluta till röstkanal
failed.voice.speak=Kunde inte börja prata
failed.voice.disconnect=Kunde inte koppla ifrån
failed.exec=Kunde inte köra
failed.fixpath=Kunde inte 'fixa' sökväg
failed.file.open=Kunde inte öppna fil
failed.file.write=Kunde inte skriva till
failed.file.read=Kunde inte läsa fil
failed.file.load=Kunde inte ladda fil
failed.file.save=Kunde inte spara fil
failed.status=Kunde inte uppdatera status
failed.typing=Kunde inte börja skriva
failed.members=Kunde inte visa medlemmarna
failed.invite.accept=Kunde inte acceptera inbjudningen
failed.invite.create=Inbjudningen kunde inte skapas
failed.roles=Kunde inte ta emot roller
failed.role.change=Kunde inte lägga till/ta bort roll
failed.role.create=Kunde inte skapa roll
failed.role.edit=Kunde inte ändra roll
failed.role.delete=Kunde inte ta bort roll!
failed.nick=Kunde inte sätta smeknamn
failed.ban.create=Kunde inte bannlysa användaren
failed.ban.delete=Kunde inte avbannlysa användaren
failed.ban.list=Kunde inte lista bannlysningar
failed.kick=Kunde inte sparka användaren
failed.leave=Kunde inte lämna
failed.block=Kunde inte blockera användaren
failed.friends=Kunde inte få vänner :(
failed.json=Kunde inte tolka JSON
failed.base64=Kunde inte konvertera till Bas64
failed.react=Could not react to message
failed.react.used=Emoji redan använd, hoppar
failed.webrequest=Det gick inte att göra webbegäran
failed.avatar=Kunde inte sätta avatar
failed.status=Kunde inte sätta status
invalid.yn=Vänligen skriv antigen 'y' eller 'n'.
invalid.webhook=Webhook-formatet är ogiltit. Format: id/token
invalid.webhook.command=Inte ett tillåtet Webhook-commando
invalid.limit.message=Meddelande överskrider teckenbegränsningen
invalid.channel=Ingen kanal vald!
invalid.guild=Ingen server vald!
invalid.value=Inget sådant värde
invalid.role=Ingen roll med det ID:t
invalid.number=Inte ett nummer
invalid.cache=Ingen cache tillgänglig!
invalid.onlyfor.users=Detta kommandot fungerar endast för användare
invalid.onlyfor.bots=Detta kommandot fungerar endast för bot-användare.
invalid.music.playing=Spelar redan något
invalid.bookmark=Bokmärket finns inte
invalid.status.offline=Offline-statusen finns men kan inte ställas in via API:n
invalid.command=Okänt kommando. Kör 'help' för att få hjälp
login.token=Vänligen klistra in en 'token' här.
login.token.user=Användar-'tokens' har prefixet 'user '
login.token.webhook=Webhook-'tokens' har prefixet 'webhook ', och deras URL eller id/token
login.starting=Autentiserar...
login.finish=Loggade in med användar-ID:t
intro.help=Kör 'help' för hjälp
intro.exit=Tryck Ctrl+D eller kör 'exit' för att avsluta.
pointer.unknown=Okänd
pointer.private=Privat
status.msg.create=Skapade meddelande med ID
status.channel=Valde kanal med ID
status.invite.accept=Accepterade inbjudan.
status.invite.create=Skapade en inbjudan med kod:
status.cache=Meddelande cache-at!
status.loading=Laddar...
status.avatar=Avatar satt!
status.name=Namn satt!
status.status=Status satt!
rl.session=Startar om session...
rl.cache.loc=Laddar om plats-cache...
rl.cache.vars=Tar bort cache-variablar...

console.=konsoll.
`
