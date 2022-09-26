# 09/11/2022

Made updates to credential logging so that it is completely unique. This was done by parsing `RIds` out of incoming `evilginx2` requests inside the `evilginx2` source code. By including a unique identifier with each credential submission, this will ensure a credential from a previous campaign will never log into a current campaign. Improvements were also made to `setup.sh`.

# 09/13/2022

Added the option to send campaign events as messages to a channel via `Microsoft Teams`.

# 09/14/2022

Removed transparency endpoint and messages from `GoPhish` altogether and increased the length of `RIds` to `10`.

# 09/16/2022

Added `SMS` campaign support with `Twilio`! Enjoy!

# 09/18/2022

Added error handling of failed `SMS` messages into the dashboard/database. If the sending of an `SMS` message fails, the error will now display right in the `GUI`. Updates to the dashboard were also made so that when viewing campaign results, statistic names now match `SMS` updates. For example, `Email Sent` now displays as `Email/SMS Sent` in the dashboard, etc.

# 09/19/2022

Added the option for operators to set their own `RId` value to use in phishing URLs to combat fingerprinting links.

# 09/20/2022

Removed `Microsoft Teams` messages to a channel and replaced with `Pusher` end-to-end encrypted channels for security reasons.

# 09/24/2022

Added notification count to `Pusher` feed app as well as sound option for events. Color changes were also made to the app :)

# 09/26/2022

Added the ability to view full token `JSON` strings from `evilginx` directly in `GoPhish` dashboard :)