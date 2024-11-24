feed_owner = "undefinedopco.de"
feed_base  = "did:plc:lbniuhsfce4bq2kqomky52px"

feed "neurodiversity" {
    name = "Bluesky Neurodiversity"
    host = "localhost"
    port = 6502
    pinned_uri = "at://did:plc:lbniuhsfce4bq2kqomky52px/app.bsky.feed.post/3lax3rm3qj22n"
    match_expr = "\\b(plural system|bipolar|bpd|schizophrenia|adhd|autism|autistic|audhd|attention deficit hyperactivity|asd|neurodiverse|neurodiversity|neurdivergent|neurospicy|neurocolorful|dyslexia|dyslexic|dyspraxic|dyspraxia|dysgraphia|dyscalculia|ocd|tourettes)\\b"

    force_expr = "\\b(#neurodiversity|#actuallyautistic|#neurodivergent)\\b"

    include_replies = true

	database = "neuro.db"

	publish "feeds.neurodifferent.me" {
        service_did = "did:web:feeds.neurodifferent.me"
        service_icon = "neurodiversity.png"
        service_short_name = "neurodiversity"
        service_human_name = "Bluesky Neurodiversity"
        service_description = "A feed showing posts with neurodiversity related terms."
	}

	exclusion_filters = [
	    "antivax",
	    "negative-terms",
	]
}

feed "disability" {
    name = "Bluesky Disability"
    host = "localhost"
    port = 6503

    match_expr = "\\b(disability|disabled|accessible|accessibility|accessible spaces|mobility aid|hearing aid|screen reader|assistive tech|visual aid|aac|tts|text to speech|assistive augmented communication|wheelchair|disabilities|ableism|ableist|impairment|disorder|access barriers|accessible design|accessible web design|accomodation|adaptive tech|braille|compensatory tool|asl|sign language|deaf|tty|teletypewriter|tbi|brain injury|universal design|sensory issues|auditory processing|apd|rsi|repetitive (stress|strain)|inclusive design|autism|autistic|adhd|eds|mcas|pots|mast cell|long covid|mecfs|dyslexia|dysgraphia|dyscalculia|mobility|neisvoid|#neisvoid|ibd|ibs|ulcerative collitis|epilepsy|ptsd|cptsd|ocd|attention deficit|chronic fatigue|cfs|me-cfs|distrophy|wheelchair|mobility scooter|#disability|#disabilityfeed)\\b"

    force_expr = "\\b(#disability|#disabilityfeed|#neisvoid|#neis|#ndis)\\b"

    include_replies = true

    database = "disability.db"

	publish "disability.neurodifferent.me" {
        service_did = "did:web:disability.neurodifferent.me"
        service_icon = "disability.png"
        service_short_name = "disability"
        service_human_name = "Bluesky Disability"
        service_description = "A feed showing posts with disability and accessibility related terms."
	}

	exclusion_filters = [
	    "antivax",
	    "negative-terms",
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

analyzer "negative-terms" {
	triggers = [
		"retard",
		"nigger",
		"faggot",	
	]

  any_trigger = true

	threshold = 0.01

	patterns = {
		"retard" = 10,
		"nigger" = 10,
		"faggot" = 10,			
	}	
}
