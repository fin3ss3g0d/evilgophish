# Table of Contents

- [evilgophish](#evilgophish)
  * [Credits](#credits)
  * [Prerequisites](#prerequisites)
  * [Disclaimer](#disclaimer)
  * [Why?](#why)
  * [Background](#background)
  * [Infrastructure Layout](#infrastructure-layout)
  * [setup.sh](#setupsh)
  * [replace_rid.sh](#replace_ridsh)
  * [Email Campaign Setup](#email-campaign-setup)
  * [SMS Campaign Setup](#sms-campaign-setup)
  * [Live Feed Setup](#live-feed-setup)
  * [Phishlets Surprise](#phishlets-surprise)
  * [A Word About Phishlets](#a-word-about-phishlets)
  * [A Note About Campaign Testing And Tracking](#a-note-about-campaign-testing-and-tracking)
  * [A Note About The Blacklist and Tracking](#a-note-about-the-blacklist-and-tracking)
  * [Changes To evilginx2](#changes-to-evilginx2)
  * [Changes to GoPhish](#changes-to-gophish)
  * [Changelog](#changelog)
  * [Issues and Support](#issues-and-support)
  * [Future Goals](#future-goals)
  * [Contributing](#contributing)

# evilgophish

Combination of [evilginx2](https://github.com/kgretzky/evilginx2) and [GoPhish](https://github.com/gophish/gophish).

## Credits

Before I begin, I would like to say that I am in no way bashing [Kuba Gretzky](https://github.com/kgretzky) and his work. I thank him personally for releasing [evilginx2](https://github.com/kgretzky/evilginx2) to the public. In fact, without his work this work would not exist. I must also thank [Jordan Wright](https://github.com/jordan-wright) for developing/maintaining the incredible [GoPhish](https://github.com/gophish/gophish) toolkit.

## Prerequisites

You should have a fundamental understanding of how to use `GoPhish`, `evilginx2`, and `Apache2`.

## Disclaimer

I shall not be responsible or liable for any misuse or illegitimate use of this software. This software is only to be used in authorized penetration testing or red team engagements where the operator(s) has(ve) been given explicit written permission to carry out social engineering. 

## Why?

As a penetration tester or red teamer, you may have heard of `evilginx2` as a proxy man-in-the-middle framework capable of bypassing `two-factor/multi-factor authentication`. This is enticing to us to say the least, but when trying to use it for social engineering engagements, there are some issues off the bat. I will highlight the two main problems that have been addressed with this project, although some other bugs have been fixed in this version which I will highlight later.

1. Lack of tracking - `evilginx2` does not provide unique tracking statistics per victim (e.g. opened email, clicked link, etc.), this is problematic for clients who want/need/pay for these statistics when signing up for a social engineering engagement.
2. Session overwriting with NAT and proxying - `evilginx2` bases a lot of logic off of remote IP address and will whitelist an IP for 10 minutes after the victim triggers a lure path. `evilginx2` will then skip creating a new session for the IP address if it triggers the lure path again (if still in the 10 minute window). This presents issues for us if our victims are behind a firewall all sharing the same public IP address, as the same session within `evilginx2` will continue to overwrite with multiple victim's data, leading to missed and lost data. This also presents an issue for our proxy setup, since `localhost` is the only IP address requesting `evilginx2`.

## Background

In this setup, `GoPhish` is used to send emails and provide a dashboard for `evilginx2` campaign statistics, but it is not used for any landing pages. Your phishing links sent from `GoPhish` will point to an `evilginx2` lure path and `evilginx2` will be used for landing pages. This provides the ability to still bypass `2FA/MFA` with `evilginx2`, without losing those precious stats. `Apache2` is simply used as a proxy to the local `evilginx2` server and an additional hardening layer for your phishing infrastructure. Realtime campaign event notifications have been provided with a local websocket/http server I have developed and full usable `JSON` strings containing tokens/cookies from `evilginx2` are displayed directly in the `GoPhish` GUI (and feed):

![new-dashboard](images/tokens-gophish.png)

## Infrastructure Layout

- `evilginx2` will listen locally on port `8443`
- `GoPhish` will listen locally on port `8080` and `3333`
- `Apache2` will listen on port `443` externally and proxy to local `evilginx2` server
  - Requests will be filtered at `Apache2` layer based on redirect rules and IP blacklist configuration
    - Redirect functionality for unauthorized requests is still baked into `evilginx2` if a request hits the `evilginx2` server

## setup.sh

`setup.sh` has been provided to automate the needed configurations for you. Once this script is run and you've fed it the right values, you should be ready to get started. Below is the setup help (note that certificate setup is based on `letsencrypt` filenames):

```
Usage:
./setup <root domain> <evilginx2 subdomain(s)> <evilginx2 root domain bool> <redirect url> <feed bool> <rid replacement>
 - root domain                     - the root domain to be used for the campaign
 - evilginx2 subdomains            - a space separated list of evilginx2 subdomains, can be one if only one
 - evilginx2 root domain bool      - true or false to proxy root domain to evilginx2
 - redirect url                    - URL to redirect unauthorized Apache requests
 - feed bool                       - true or false if you plan to use the live feed
 - rid replacement                 - replace the gophish default "rid" in phishing URLs with this value
Example:
  ./setup.sh example.com "accounts myaccount" false https://redirect.com/ true user_id
```

Redirect rules have been included to keep unwanted visitors from visiting the phishing server as well as an IP blacklist. The blacklist contains IP addresses/blocks owned by ProofPoint, Microsoft, TrendMicro, etc. Redirect rules will redirect known *"bad"* remote hostnames as well as User-Agent strings. 

## replace_rid.sh

In case you ran `setup.sh` once and already replaced the default `RId` value throughout the project, `replace_rid.sh` was created to replace the `RId` value again.

```
Usage:
./replace_rid <previous rid> <new rid>
 - previous rid      - the previous rid value that was replaced
 - new rid           - the new rid value to replace the previous
Example:
  ./replace_rid.sh user_id client_id
```

## Email Campaign Setup

Once `setup.sh` is run, the next steps are: 

1. Start `GoPhish` and configure email template, email sending profile, and groups
2. Start `evilginx2` and configure phishlet and lure (must specify full path to `GoPhish` `sqlite3` database with `-g` flag)
3. Ensure `Apache2` server is started
4. Launch campaign from `GoPhish` and make the landing URL your lure path for `evilginx2` phishlet
5. **PROFIT**

## SMS Campaign Setup

An entire reworking of `GoPhish` was performed in order to provide `SMS` campaign support with `Twilio`. Your new `evilgophish` dashboard will look like below:

![new-dashboard](images/new-dashboard.png)

Once you have run `setup.sh`, the next steps are:

1. Configure `SMS` message template. You will use `Text` only when creating a `SMS` message template, and you should not include a tracking link as it will appear in the `SMS` message. Leave `Envelope Sender` and `Subject` blank like below:

![sms-message-template](images/sms-message-template.png)

2. Configure `SMS Sending Profile`. Enter your phone number from `Twilio`, `Account SID`, `Auth Token`, and delay in between messages into the `SMS Sending Profiles` page:

![sms-sending-profile](images/sms-sending-profile.png)

3. Import groups. The `CSV` template values have been kept the same for compatibility, so keep the `CSV` column names the same and place your target phone numbers into the `Email` column. Note that `Twilio` accepts the following phone number formats, so they must be in one of these three:

![twilio-number-formats](images/twilio-number-formats.png)

4. Start `evilginx2` and configure phishlet and lure (must specify full path to `GoPhish` `sqlite3` database with `-g` flag)
5. Ensure `Apache2` server is started
6. Launch campaign from `GoPhish` and make the landing URL your lure path for `evilginx2` phishlet
7. **PROFIT**

## Live Feed Setup

Realtime campaign event notifications are handled by a local websocket/http server and live feed app. To get setup:

1. Select `true` for `feed bool` when running `setup.sh`

2. `cd` into the `evilfeed` directory and start the app with `./evilfeed`

3. When starting `evilginx2`, supply the `-feed` flag to enable the feed. For example:

`./evilginx2 -feed -g /opt/evilgophish/gophish/gophish.db`

4. You can begin viewing the live feed at: `http://localhost:1337/`. The feed dashboard will look like below:

![live-feed](images/live-feed.png)

**IMPORTANT NOTES**

- The live feed page hooks a websocket for events with `JavaScript` and you **DO NOT** need to refresh the page. If you refresh the page, you will **LOSE** all events up to that point.

## Phishlets Surprise

Included in the `evilginx2/phishlets` folder are three custom phishlets not included in [evilginx2](https://github.com/kgretzky/evilginx2). 

1. `o3652` - modified/updated version of the original `o365` (stolen from [Optiv blog](https://www.optiv.com/insights/source-zero/blog/spear-phishing-modern-platforms))
2. `google` - updated from previous examples online (has issues since release, don't use in live campaigns)
3. `knowbe4` - custom ([demo](https://youtu.be/iDxFpcdXddU))

## A Word About Phishlets

I feel like the world has been lacking some good phishlet examples lately. It would be great if this repository could be a central repository for the latest phishlets. Send me your phishlets at `fin3ss3g0d@pm.me` for a chance to end up in `evilginx2/phishlets`. If you provide quality work, I will create a `Phishlets Hall of Fame` and you will be added to it.

## A Note About Campaign Testing And Tracking

It is not uncommon to test the tracking for a campaign before it is launched and I encourage you to do so, I will just leave you with a warning. `evilginx2` will create a cookie and establish a session for each new victim's browser. If you continue to test multiple campaigns and multiple phishing links within the same browser, you will confuse the tracking process since the `RId` value is parsed out of requests and set at the start of a new session. If you are doing this, you are not truly simulating a victim as a victim would never have access to another phishing link besides their own and goes without saying that this will never happen during a live campaign. This is to fair warn you not to open an issue for this as you are not using the tool the way it was intended to be used. If you would like to simulate a new victim, you can test the tracking process by using a new browser/tab in incognito mode.

## A Note About The Blacklist and Tracking

As mentioned above, there is an IP address blacklist included in this project that may cause some clients to get blocked and disrupt the tracking process. For right now, it is up to you to perform test campaigns and verify if any blocking will disrupt your campaign tracking. A blocked client will receive a `403 Forbidden` error. `/var/log/apache2/access_evilginx2.log` can be viewed for remote IP addresses accessing the phishing server. You can remove entries in the `/etc/apache2/blacklist.conf` file that are causing a tracking issue and restart Apache. Or you can remove the `Location` block in the `/etc/apache2/sites-enabled/000-default.conf` file and restart Apache to remove IP blacklisting altogether.

## Changes To evilginx2

1. All IP whitelisting functionality removed, new proxy session is established for every new visitor that triggers a lure path regardless of remote IP
2. Fixed issue with phishlets not extracting credentials from `JSON` requests
3. Further *"bad"* headers have been removed from responses
4. Added logic to check if `mime` type was failed to be retrieved from responses
5. All `X` headers relating to `evilginx2` have been removed throughout the code (to remove IOCs)
6. Added phishlets

## Changes to GoPhish

1. All `X` headers relating to `GoPhish` have been removed throughout the code (to remove IOCs)
2. Default `rid` string in phishing URLs is chosen by the operator in `setup.sh`
3. Added `SMS` Campaign Support

## Changelog 

See the `CHANGELOG.md` file for changes made since the initial release.

## Issues and Support

I am mostly looking for legitimate bugs in code or enhancement opportunities and not to be a personal help desk support for struggles during your setup. You should understand the prerequisites of setting up a social engineering campaign including how `Apache`, `DNS`, SSL certificates, `evilginx2`, `gophish`, and proxies work to use and setup this tool. With that being said, issues falling into these categories will be closed. I am taking the same stance as [Kuba Gretzky](https://github.com/kgretzky) and will not help creating phishlets. There are plenty of examples of working phishlets and for you to create your own, if you open an issue for a phishlet it will be closed. However, I *will* maintain *certain* phishlets at will (see [A Word About Phishlets](#a-word-about-phishlets)). I will state for the record that tracking for this project works as advertised and if it does not, it is a result of a misconfiguration during `setup.sh` or you are confusing the tool by visiting multiple `RId`s within the same browser session. Do not open an issue for this or it will be closed (see [A Note About Campaign Testing And Tracking](#a-note-about-campaign-testing-and-tracking)). If you open an issue, please provide as much detailed information as possible about the issue including output pertaining to the issue. Issues with lack of detail or output will be closed. 

## Future Goals

- Test/review/pull `evilginx3` update
- Additions to IP blacklist and redirect rules
- Add more phishlets

## Contributing

I would like to see this project improve and grow over time. If you have improvement ideas, new redirect rules, new IP addresses/blocks to blacklist, phishlets, or suggestions, please email me at: `fin3ss3g0d@pm.me` or open a pull request.