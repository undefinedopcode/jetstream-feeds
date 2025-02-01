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