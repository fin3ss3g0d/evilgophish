# evilgophish

Combination of [evilginx2](https://github.com/kgretzky/evilginx2) and [GoPhish](https://github.com/gophish/gophish).

## Credits

Before I begin, I would like to say that I am in no way bashing [Kuba Gretzky](https://github.com/kgretzky) and his work. I thank him personally for releasing [evilginx2](https://github.com/kgretzky/evilginx2) to the public. In fact, without his work this work would not exist. I must also thank [Jordan Wright](https://github.com/jordan-wright) for developing/maintaining the incredible [GoPhish](https://github.com/gophish/gophish) toolkit. The last thank you I must make is to [Fotios Liatsis](https://twitter.com/_wizard32?lang=en) and [Outpost24](https://outpost24.com/) for sharing their solution to combine these two frameworks.

## Prerequisites

You should have a fundamental understanding of how to use `GoPhish`, `evilginx2`, and `Apache2`.

## Why?

As a penetration tester or red teamer, you may have heard of `evilginx2` as a proxy man-in-the-middle framework capable of bypassing `two-factor/multi-factor authentication`. This is enticing to us to say the least, but when trying to use it for social engineering engagements, there are some issues off the bat. I will highlight the two main problems that have been addressed with this project, although some other bugs have been fixed in this version which I will highlight later.

1. Lack of tracking - `evilginx2` does not provide unique tracking statistics per victim (e.g. opened email, clicked link, etc.), this is problematic for clients who want/need/pay for these statistics when signing up for a social engineering engagement.
2. Session overwriting with NAT and proxying - `evilginx2` bases a lot of logic off of remote IP address and will whitelist an IP for 10 minutes after the victim triggers a lure path. `evilginx2` will then skip creating a new session for the IP address if it triggers the lure path again (if still in the 10 minute window). This presents issues for us if our victims are behind a firewall all sharing the same public IP address, as the same session within `evilginx2` will continue to overwrite with multiple victim's data, leading to missed and lost data. This also presents an issue for our proxy setup, since `localhost` is the only IP address requesting `evilginx2`.

## How Does It Work?

This project is based off of this [blog](https://outpost24.com/blog/Better-proxy-than-story), so please read it before getting started. To provide the tracking part, there is a function that runs every minute inside of `GoPhish` that keeps refreshing the dashboard with results for a campaign. This function was modified to loop all sent email events and for each, loop the `Apache2` access log for `evilginx2` looking for a regular expression match for the `RId` associated with that sent email event. If a match is found, this indicates a user clicked their link and a clicked link event for that user is created inside `GoPhish`. From there, all clicked link events are looped and a custom credential log from `evilginx2` is looped looking for a regular expression match for the username inside of `GoPhish`. If there is a match here, we know the user submitted data and a submitted data event is created inside `GoPhish` containing the username/password combination from `evilginx2`.

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

## Starting A Campaign

You will proceed to send emails from `GoPhish` as normal, however use a fake landing and instead your phishing link will point to your `evilginx2` lure. `GoPhish` will automatically add a `RId` to the end of it for tracking.

### Ensuring Email Opened Tracking

You **CANNOT** use the default `Add Tracking Image` button when creating your email template. You **MUST** include your own image tag that points at the `GoPhish` server with the tracking URL scheme. This is also explained/shown in the [blog](https://outpost24.com/blog/Better-proxy-than-story).

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

## Issues and Support

I am taking the same stance as [Kuba Gretzky](https://github.com/kgretzky) and will not help creating phishlets. There are plenty of examples of working phishlets and for you to create your own, if you open an issue for a phishlet it will be closed. I will also not consider issues with your `Apache2`, `DNS`, or certificate setup as legitimate issues and they will be closed. Please read the included [blog](https://outpost24.com/blog/Better-proxy-than-story) for how to get setup properly. However, if you encounter a legitimate failure/error with the program, I will take the issue seriously.

## Contributing

I would like to see this project improve and grow over time. If you feel you have improvement opportunities or ideas, please email me at: `fin3ss3g0d@pm.me`.