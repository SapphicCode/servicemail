# `servicemail`

Servicemail is a service mail ingestion and delivery service. It's designed to
deal with those annoying `noreply` email addresses that would otherwise clutter
your inbox, and instead route them to a messaging service of your choice.

## Premise

The goal of Servicemail is to allow you to create accounts with services under
random email addresses with no correlation to one another. This makes the job
of hackers a nightmare, because they can't just cross-reference your login 
credentials from a password breach somewhere. Even better, if your address
gets leaked to a spam database, just revoke it, and the address will stop
routing messages.

As mentioned, Servicemail will proxy all of the email coming to a given address
directly to your favorite instant messenger of choice. Since the primary way
of replying to these emails is either entering a code in the requesting app, or
opening a link inside the email, this solution fits perfectly (and stops those
emails from otherwise cluttering up your inbox).

The goal of Servicemail is **not** to let you breach services' Terms of
Service. That is to say, if a service requires your email for personal
interaction, or only lets you create a single account within the bounds of
their ToS, you may not use Servicemail to bypass that restriction.
