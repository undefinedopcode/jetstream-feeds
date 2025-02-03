## Jetstream Feeds

`jetstream-feeds` is a bluesky feed daemon designed for hosting feeds based on consuming jetstream. 

### Configuration

Config is performed using a `feeds.hcl` file using HCL (HashiCorp Configuration Language).

#### Global config

```hcl
feed_owner = "foobarbaz@bsky.social"
feed_base  = "did:plc:fj234r9gj345jm340fgm"
```

- `feed_owner` is the human readable handle of the feed owner.
- `feed_base` is the DID of the feed owner.

#### Feed config

A feed block defines a feed instance running on a particular port. It may specify a pinned post via `pinned_uri` that can provide a description of what the feed is about.

- `match_expr` is a regexp (go compatible) used to test posts for a match.
- `force_expr` is an optional regexp for which a match will forcibly include the post.
- `include_replies` specifies whether a post reply may be included in the feed if it matches, versus just an initial post in a thread.
- `database` names an sqlite3 database to use for storing feed uris that match.
- The `publish` block defines information for publishing the feed to bluesky (which may be done with the `-publish <feed>` command line flag)
  - `service_did` provides a unique indentifier for the feed. Typically based on it's web address.
  - `service_icon` provides an optional png icon for the feed. 
  - `service_short_name` provides a short name for the feed uri in bluesky (should be unique under your identity)
  - `service_human_name` is the name you would like to use in the bluesky feeds list.
  - `service_description` provides a sentence summarizing what the feed is about for the bluesky feeds list.
- `exclusion_filters` may list one or more filters that will exclude posts based on simple scoring of the post content.

```hcl
feed "ducks" {
    name = "Bluesky Duck fanciers"
    host = "localhost"
    port = 6502
    pinned_uri = "at://did:plc:lbniuhsfce4bq2kqomky52px/app.bsky.feed.post/3lax3rm3qj22n"

    match_expr = "\\b(ducks|quack|canard)\\b"

    force_expr = "\\b(#ducks)\\b"

    include_replies = true

	database = "ducks.db"

	publish "ducks.example.com" {
        service_did = "did:web:ducks.example.com"
        service_icon = "ducks.png"
        service_short_name = "ducks"
        service_human_name = "Bluesky Duck fanciers"
        service_description = "A feed showing posts with duck related terms."
	}

	exclusion_filters = [
	    "antivax",
	]
}
```

#### Exclusion filters (simple sentiment filter)

Defines exclusion filters that may be used in multiple feeds. Each has a list of potential `trigger` words that flag the post for checking by the filter.

`threshold` defines a score value for which if a post scopes equal or higher to, it will be excluded even if the post matches `match_expr` or `force_expr`.

`patterns` defines a list of strings, each mapped to a score value. Terms which increase the posts exclusion likelyhood map to a postive float.  Terms which reduce the posts likelyhood of exclusion map to a negative float. 

Below is an example filter, developed on antivax posts from 'X'.  This filter is live on the bsky neurodiversity feed, as users had encountered a large amount of antivax posts appearing in the feed. 

```hcl
analyzer "antivax" {
    triggers = [
        "autism",
        "vaccine"
    ]

    threshold = 0.4

    patterns = {
        "misinformation" =                     0.6,
        "conspiracy" =                         0.7,
        "dangerous" =                          0.5,
        "unsafe" =                             0.5,
        "unproven" =                           0.5,
        "experimental" =                       0.4,
        "cause autism" =                       0.2,
        "do cause autism" =                    0.6,
        "don't cause autism" =                -1.0,
        "poisoning" =                          0.5,
        "toxins" =                             0.5,
        "don't think vaccines cause autism" = -0.5,
        "do not cause autism" =               -0.5,
        "did not cause my childs autism" =    -0.5,
        "big pharma" =                         0.7,
        "too many vaccines" =                  0.7,
        "vaccine industry" =                   0.7,
        "vaccines destroy" =                   0.8,
        "vaccine induced" =                    0.9,
        "wakefield" =                          0.3,
        "kirsch" =                             0.3,
        "corrupt" =                            0.3,
        "cdc is lying" =                       0.3,
        "cdc has been lying" =                 0.3,
        "medical community claims" =           0.2,
    }
}
```

### Publishing a feed

You can publish a feed via the -publish command line option. 

```sh
./jetstream-feeds -publish "ducks"
```

You should create a single use app-password on bluesky, and you will be prompted for this password. 

Enter it, and the feed (if properly configured should be published).

#### Hosting the published feed

The feeds run on a localhost port with no encryption and are designed to have a webserver like nginx sit in front of it. This allows specifying a custom domain, and providing a certificate forn the host, as bluesky expects a valid HTTP certificate to be provided.  

A LetsEncrypt certificate will be sufficient for this purpose. 

Complex example config:

```hcl
feed_owner = "foobarbaz@bsky.social"
feed_base  = "did:plc:fj234r9gj345jm340fgm"

feed "ducks" {
    name = "Bluesky Duck fanciers"
    host = "localhost"
    port = 6502
    pinned_uri = "at://did:plc:lbniuhsfce4bq2kqomky52px/app.bsky.feed.post/3lax3rm3qj22n"

    match_expr = "\\b(ducks|quack|canard)\\b"

    force_expr = "\\b(#ducks)\\b"

    include_replies = true

	database = "ducks.db"

	publish "ducks.example.com" {
        service_did = "did:web:ducks.example.com"
        service_icon = "ducks.png"
        service_short_name = "ducks"
        service_human_name = "Bluesky Duck fanciers"
        service_description = "A feed showing posts with duck related terms."
	}

	exclusion_filters = [
	    "antivax",
	]
}

analyzer "antivax" {
    triggers = [
        "autism",
        "vaccine"
    ]

    threshold = 0.4

    patterns = {
        "misinformation" =                     0.6,
        "conspiracy" =                         0.7,
        "dangerous" =                          0.5,
        "unsafe" =                             0.5,
        "unproven" =                           0.5,
        "experimental" =                       0.4,
        "cause autism" =                       0.2,
        "do cause autism" =                    0.6,
        "don't cause autism" =                -1.0,
        "poisoning" =                          0.5,
        "toxins" =                             0.5,
        "don't think vaccines cause autism" = -0.5,
        "do not cause autism" =               -0.5,
        "did not cause my childs autism" =    -0.5,
        "big pharma" =                         0.7,
        "too many vaccines" =                  0.7,
        "vaccine industry" =                   0.7,
        "vaccines destroy" =                   0.8,
        "vaccine induced" =                    0.9,
        "wakefield" =                          0.3,
        "kirsch" =                             0.3,
        "corrupt" =                            0.3,
        "cdc is lying" =                       0.3,
        "cdc has been lying" =                 0.3,
        "medical community claims" =           0.2,
    }
}
```