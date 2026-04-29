package telegram

const msgHelp = `I can save and keep your links. I also send you a random saved link after every configured number of saved links. 😁

To save a link, send me a full http:// or https:// URL. 😌

To get a random saved link manually, send /rnd.`

const msgHello = "Hi there! 👾\n\n" + msgHelp

const (
	msgUnknownCommand = "Unknown command 🤔"
	msgNoSavedPages   = "You have no saved pages 🙊"
	msgSaved          = "Saved! 👌"
	msgAlreadyExists  = "You have already have this page in your list 🤗"
	msgInvalidURL     = "Please send a valid http:// or https:// URL."
)
