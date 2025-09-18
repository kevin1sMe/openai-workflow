# <img src='Workflow/icon.png' width='45' align='center' alt='icon'> ChatGPT / DALL-E Alfred Workflow

OpenAI integrations

[⤓ Install on the Alfred Gallery](https://alfred.app/workflows/alfredapp/openai)

## Setup

1. Create an OpenAI account and [log in](https://platform.openai.com/login?launch).
2. On the [API keys page](https://platform.openai.com/api-keys), click `+ Create new secret key`.
3. Name your new secret key and click `Create secret key`.
4. Copy your secret key and add it to the [Workflow’s Configuration](https://www.alfredapp.com/help/workflows/user-configuration/).

## Storage Security

Set the `storage_secret` workflow variable to enable at-rest encryption for chat history and the streaming state. The secret is hashed with SHA-256 and used with AES-GCM, so pick a long, unique passphrase.

* Existing plain-text histories are re-encrypted the next time you send a message.
* `chat.json` and `stream.txt` are written with `0600` permissions once encryption is on.
* Leave the variable empty if you prefer the previous plain-text behaviour.

## Usage

### ChatGPT

Query ChatGPT via the `chatgpt` keyword, the [Universal Action](https://www.alfredapp.com/help/features/universal-actions/), or the [Fallback Search](https://www.alfredapp.com/help/features/default-results/fallback-searches/).

![Start ChatGPT query](Workflow/images/about/chatgptkeyword.png)

![Querying ChatGPT](Workflow/images/about/chatgpttextview.png)

* <kbd>↩</kbd> Ask a new question.
* <kbd>⌘</kbd><kbd>↩</kbd> Clear and start new chat.
* <kbd>⌥</kbd><kbd>↩</kbd> Copy last answer.
* <kbd>⌃</kbd><kbd>↩</kbd> Copy full chat.
* <kbd>⇧</kbd><kbd>↩</kbd> Stop generating answer.

#### Chat History

View Chat History with ⌥↩ in the `chatgpt` keyword. Each result shows the first question as the title and the last as the subtitle.

![Viewing chat histories](Workflow/images/about/chatgpthistory.png)

<kbd>↩</kbd> to archive the current chat and load the selected one. Older chats can be trashed with the `Delete` [Universal Action](https://www.alfredapp.com/help/features/universal-actions/). Select multiple chats with the [File Buffer](https://www.alfredapp.com/help/features/file-search/#file-buffer).

### DALL·E

Query DALL·E via the `dalle` keyword.

![Start DALL-E query](Workflow/images/about/dallekeyword.png)

![Querying DALL-E](Workflow/images/about/dalletextview.png)

* <kbd>↩</kbd> Send a new prompt.
* <kbd>⌘</kbd><kbd>↩</kbd> Archive images.
* <kbd>⌥</kbd><kbd>↩</kbd> Reveal last image in the Finder.
