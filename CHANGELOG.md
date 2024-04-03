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

# 09/30/2022

Removed the need to create a fake landing inside `GoPhish` as part of a campaign setup. Removed `Add Tracking Image` button to avoid mistakes being made with operators using this button. This forces the manual insertion of image tags into email templates as it is needed for tracking to properly work. Also added enhancements to the logging process.

# 10/02/2022

Added the `Add Tracking Image` button back and ability to handle email opened tracking events directly from `evilginx2`. This removes the need to proxy to the local `GoPhish` server altogether.

# 10/03/2022

Made further enhancements to the logging process so that if a user triggers an event more than once, every instance will now be captured. Changed `Pusher` notifications to trigger directly from `evilginx` to ensure feed is completely realtime.

# 10/05/2022

Removed `Pusher` messages to a channel to avoid message limit and to ease the live feed setup process. The live feed has been changed to operate completely local, without sending victim data to any remote `API`. The live feed also now has no message limits and added the ability to view full token JSON in live feed :)

# 10/29/2022

Made `Apache2` blacklist optional.

# 01/10/2023

Added `Cisco VPN` phishlet, merged pull request that allows operators to get source `IP` information for victims when generating `GoPhish` reports.

# 02/03/2023

Added some improved logic for logging credentials to `GoPhish` where sometimes the username parameter of a phishlet was lost due to not checking if it was empty. This should improve the overall user experience and credential logging.

# 03/14/2023

Removed a "X-Evilginx" header IOC that was hidden as a XOR encrypted byte array.

# 04/09/2023

Added the option to force a Google reCAPTCHA v2 challenge before granting access to a lure. This was done to combat domain takedowns and thwart off bots.

# 04/19/2023

Added the option to force a Cloudflare Turnstile challenge before granting access to lure. Again this was done to try to thwart off bots.

# 07/08/2023

Upgraded `evilginx2` to `evilginx3`! :)

# 07/08/2023

Deprecated the `Cloudflare Turnstile` and `Google reCAPTCHA` features due to the `redirectors` feature of `evilginx3` which accomplishes the same task.

# 07/15/2023

Pulled the `evilginx3` version `3.1.0` update so now all cookie tokens are captured by default. Removed the need to input a certificate path when running `setup.sh` since it is handled for users now with `evilginx3`. Removed a small bug inside of `gophish` that would attempt to post a notification to the live feed for when emails are sent despite a user specifying not to use the feed.

# 09/23/2023

Updated `evilginx3` to version `3.2.0`. Fixed an issue with `gophish` not being able to add custom email headers.

# 02/23/2024

Added QR code generator feature allowing operators to deploy QR code social engineering campaigns. This was email only, SMS capability is being researched for the future.

# 03/01/2024

Added `evilginx3` commits [kgretzky/evilginx2@d8f7d44e1450e8673a4a78e77c8041de12a02229](https://github.com/kgretzky/evilginx2/commit/d8f7d44e1450e8673a4a78e77c8041de12a02229), [kgretzky/evilginx2@3b0f5c9971bf1041acc88d1b6ffcb9a5203f261c](https://github.com/kgretzky/evilginx2/commit/3b0f5c9971bf1041acc88d1b6ffcb9a5203f261c), [kgretzky/evilginx2@e7a68662a02a83fbf2c2c4914f46d191d3952ed1](https://github.com/kgretzky/evilginx2/commit/e7a68662a02a83fbf2c2c4914f46d191d3952ed1), & [kgretzky/evilginx2@1b9cb590fefcf30d2f6a460e17098b43182d3c4f](https://github.com/kgretzky/evilginx2/commit/1b9cb590fefcf30d2f6a460e17098b43182d3c4f) which fix various issues, introduce some code cleanup, and add new features!

# 03/04/2024

Added enhancements to the SMS sending process, allowing users to specify a launch date and send by date, removing the delay feature.

# 03/11/2024

Changed the rID generator algorithm so that not all rID values are a set character length.

# 03/30/2024

Added `evilginx3` commit [kgretzky/evilginx2@edadd5233907985f550e11f144924b1a1882f944](https://github.com/kgretzky/evilginx2/commit/edadd5233907985f550e11f144924b1a1882f944) adding support for JSON force post functionality.

# 04/03/2024

Removed `Gophish` email transparency headers.