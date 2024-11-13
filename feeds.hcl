feed_owner = "undefinedopco.de"
feed_base  = "did:plc:lbniuhsfce4bq2kqomky52px"

feed "neurodiversity" {
    name = "Bluesky Neurodiversity"
    port = 6502
    match_expr = "\\b(plural system|bipolar|bpd|schizophrenia|adhd|autism|autistic|audhd|attention deficit hyperactivity|asd|neurodiverse|neurodiversity|neurdivergent|neurospicy|neurocolorful|dyslexia|dyslexic|dyspraxic|dyspraxia|dysgraphia|dyscalculia|ocd|tourettes)\\b"
    force_expr = "\\b(#neurodiversity|#actuallyautistic)\\b"
    include_replies = false
	database = "neuro.db"
}

feed "disability" {
    name = "Bluesky Disability"
    port = 6503
    match_expr = "\\b(disability|disabled|accessible|accessibility|accessible spaces|mobility aid|hearing aid|screen reader|assistive tech|visual aid|aac|tts|text to speech|assistive augmented communication|wheelchair|disabilities|ableism|ableist|impairment|disorder|access barriers|accessible design|accessible web design|accomodation|adaptive tech|braille|compensatory tool|asl|sign language|deaf|blind|tty|teletypewriter|tbi|brain injury|universal design|sensory issues|auditory processing|apd|rsi|repetitive (stress|strain)|inclusive design|autism|autistic|adhd|eds|mcas|pots|mast cell|long covid|mecfs|dyslexia|dysgraphia|dyscalculia|mobility|neisvoid|#neisvoid|ibd|ibs|ulcerative collitis|epilepsy|ptsd|cptsd|ocd|attention deficit|chronic fatigue|cfs|me-cfs|distrophy|wheelchair|mobility scooter|#disability|#disabilityfeed)\\b"
    force_expr = "\\b(#disability|#disabilityfeed)\\b"
    include_replies = false
    database = "disability.db"
}
