# evilgophish

Combination of [evilginx2](https://github.com/kgretzky/evilginx2) and [GoPhish](https://github.com/gophish/gophish).

## Credits

Before I begin, I would like to say that I am in no way bashing [Kuba Gretzky](https://github.com/kgretzky) and his work. I thank him personally for releasing [evilginx2](https://github.com/kgretzky/evilginx2) to the public. In fact, without his work this work would not exist. I must also thank [Jordan Wright](https://github.com/jordan-wright) for developing/maintaining the incredible [GoPhish](https://github.com/gophish/gophish) toolkit. The last thank you I must make is to [Fotios Liatsis](https://twitter.com/_wizard32?lang=en) and [Outpost24](https://outpost24.com/) for sharing their solution to combine these two frameworks.

## Prerequisites

You should have a fundamental understanding of how to use `GoPhish`, `evilginx2`, and `Apache2`.

## Disclaimer

I shall not be responsible or liable for any misuse or illegitimate use of this software. This software is only to be used in authorized penetration testing or red team engagements where the operator(s) has(ve) been given explicit written permission to carry out social engineering. 

## Why?

As a penetration tester or red teamer, you may have heard of `evilginx2` as a proxy man-in-the-middle framework capable of bypassing `two-factor/multi-factor authentication`. This is enticing to us to say the least, but when trying to use it for social engineering engagements, there are some issues off the bat. I will highlight the two main problems that have been addressed with this project, although some other bugs have been fixed in this version which I will highlight later.

1. Lack of tracking - `evilginx2` does not provide unique tracking statistics per victim (e.g. opened email, clicked link, etc.), this is problematic for clients who want/need/pay for these statistics when signing up for a social engineering engagement.
2. Session overwriting with NAT and proxying - `evilginx2` bases a lot of logic off of remote IP address and will whitelist an IP for 10 minutes after the victim triggers a lure path. `evilginx2` will then skip creating a new session for the IP address if it triggers the lure path again (if still in the 10 minute window). This presents issues for us if our victims are behind a firewall all sharing the same public IP address, as the same session within `evilginx2` will continue to overwrite with multiple victim's data, leading to missed and lost data. This also presents an issue for our proxy setup, since `localhost` is the only IP address requesting `evilginx2`.

## General Background

This project is based on this [blog](https://outpost24.com/blog/Better-proxy-than-story) and I encourage you to read it before getting started. In this setup, `GoPhish` is used to send emails, track opened emails, and provide a dashboard for `evilginx2` campaign statistics, but it is not used for any landing pages. To provide tracking between the two, the function resposible for providing campaign results inside `GoPhish` has been modified to instead get clicked link event details and submitted data event details from logs related to `evilginx2`. Your phishing link sent from `GoPhish` will point to an `evilginx2` lure path and `evilginx2` will be used for landing pages. This provides the ability to still bypass `2FA/MFA` with `evilginx2`, without losing those precious stats. The operator will also be informed of a submitted data event in realtime. This should ensure the operator won't not run out of time to use captured cookies, or at least be informed as soon as possible. The operator will still need to bounce over to the `evilginx2` terminal to fetch the full `JSON` string of captured tokens/cookies.

## Infrastructure Layout

- `evilginx2` will listen locally on port `8443`
- `GoPhish` will listen locally on port `8080`
- `Apache2` will listen on port `443` externally and proxy to either local `GoPhish/evilginx2` depending on the subdomain name requested. `Apache2` access log file is created for both `GoPhish/evilginx2` servers
  - Requests will be filtered at `Apache2` layer based on redirect rules and IP blacklist configuration

## Getting Setup

Assuming you have read the [blog](https://outpost24.com/blog/Better-proxy-than-story) and understand how the setup works, `setup.sh` has been provided to automate the needed configurations for you. Once this script is run and you've fed it the right values, you should be ready to get started. Below is the setup help (note that certificate setup is based on `letsencrypt` filenames):

```
Usage:
./setup <root domain> <evilginx2 subdomain(s)> <gophish subdomain(s)> <redirect url>
 - root domain             - the root domain to be used for the campaign
 - evilginx2 subdomains    - a space separated list of evilginx2 subdomains, can be one if only one
 - gophish subdomains      - a space separated list of gophish subdomains, can be one if only one
 - redirect url            - URL to redirect unauthorized Apache requests
Example:
  ./setup.sh example.com "training login" "download www" https://redirect.com/
```

Redirect rules have been included to keep unwanted visitors from visiting the phishing server as well as an IP blacklist. The blacklist contains IP addresses/blocks owned by ProofPoint, Microsoft, TrendMicro, etc. Redirect rules will redirect known *"bad"* remote hostnames as well as User-Agent strings. 

Once the setup script is run, the next steps are: 

1. Make sure the `Apache2` log file for `evilginx2` exists before starting `GoPhish` (starting `Apache2` will automatically do this)
2. Start `GoPhish` and configure email template (see note below about email opened tracking), email sending profile, fake landing page, and groups
3. Start `evilginx2` and configure phishlet and lure
4. Launch campaign from `GoPhish` and make the landing URL your lure path for `evilginx2` phishlet
5. **PROFIT**

## Ensuring Email Opened Tracking

You **CANNOT** use the default `Add Tracking Image` button when creating your email template. You **MUST** include your own image tag that points at the `GoPhish` server with the tracking URL scheme. This is also explained/shown in the [blog](https://outpost24.com/blog/Better-proxy-than-story). For example, if your `GoPhish` subdomain is `download.example.org`, and your `evilginx2` lure path is `https://login.example.org/login`, you would include the following tag in your email `.html` which will provide email opened tracking in `GoPhish`:

`<img src="https://download.example.org/login/track?client_id={{.RId}}"/>`

## Changes To evilginx2

1. All IP whitelisting functionality removed, new proxy session is established for every new visitor that triggers a lure path regardless of remote IP
2. Custom credential logging on submitted passwords to `~/.evilginx/creds.json`
3. Fixed issue with phishlets not extracting credentials from `JSON` requests
4. Further *"bad"* headers have been removed from responses
5. Added logic to check if `mime` type was failed to be retrieved from responses
6. All `X` headers relating to `evilginx2` have been removed throughout the code (to remove IOCs)

## Changes to GoPhish

1. Custom logic inserted into `GetCampaignResults` function that handles `evilginx2` tracking from Apache2 access log
2. Custom logging of events to `JSON` format in `HandleEvent` functions
3. Additional config parameter added for Apache2 log path
4. All `X` headers relating to `GoPhish` have been removed throughout the code (to remove IOCs)
5. Default server name has been changed to `IGNORE`
6. Custom 404 page functionality, place a `.html` file named `404.html` in `templates` folder (example has been provided)
7. `rid=` is now `client_id=` in phishing URLs

## Phishlets Surprise

Included in the `evilginx2/phishlets` folder are three custom phishlets not included in [evilginx2](https://github.com/kgretzky/evilginx2). 

1. `O3652` - modified/updated version of the original `o365` (stolen from [Optiv blog](https://www.optiv.com/insights/source-zero/blog/spear-phishing-modern-platforms))
2. `google` - updated from previous examples online
3. `knowbe4` - custom (don't have access to an account for testing auth URL, works for single-factor campaigns, have not fully tested MFA)

## Limitations 

- All events will only be submitted once into `GoPhish`
- If you do multiple campaigns targeting the same victims without deleting `~/.evilginx/creds.json`, credentials from a previous campaign will take presedence in `GoPhish`

## **Important Notes**

You **MUST** make sure `Apache2` is logging to the file defined in `gophish/config.json` for the `evilginx2` server, the default path is `/var/log/apache2/access_evilginx2.log` unless you change it. For example, if `Apache2` is logging to `/var/log/apache2/access_evilginx2.log.1` and you have `/var/log/apache2/access_evilginx2.log` defined in `gophish/config.json`, you will lose tracking statistics. All credentials ever captured by `evilginx2` will log to `~/.evilginx/creds.json`. If you use your server for multiple campaigns that may end up targeting the same victims, you will want to make a backup of it and then delete it to prevent previous credentials from logging into current campaigns.

## Issues and Support

I am taking the same stance as [Kuba Gretzky](https://github.com/kgretzky) and will not help creating phishlets. There are plenty of examples of working phishlets and for you to create your own, if you open an issue for a phishlet it will be closed. I will also not consider issues with your `Apache2`, `DNS`, or certificate setup as legitimate issues and they will be closed. Please read the included [blog](https://outpost24.com/blog/Better-proxy-than-story) for how to get setup properly. However, if you encounter a legitimate failure/error with the program, I will take the issue seriously.

## Future Goals

- `Microsoft Teams` notifications to a channel upon submitted credentials (this will most likely happen, stay tuned!)
- Additions to IP blacklist and redirect rules

## Contributing

I would like to see this project improve and grow over time. If you have improvement ideas, new redirect rules, new IP addresses/blocks to blacklist, or suggestions, please email me at: `fin3ss3g0d@pm.me`.