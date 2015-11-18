Logopoeist
==========

Logopoeist is a random word-generation system for conlangers (creators of constructed languages) which conforms to user-specified phonotactic rules as well as user-specified conditional probability distributions to control phoneme frequencies.

Pre-compiled binaries for 64-bit Windows, Mac OS, and Linux can be downloaded on the release page: https://github.com/conlang-software-dev/Logopoeist/releases/tag/v0.1

Logopoeist is a command-line program, and takes the following arguments:

* `-file {string}`: the name of an input configuration file. If absent, Logopoeist will try to read configuration commands from standard input.
* `-n {int}`: the number of random words to output. Defaults to 10.

A sample configuration for a strict-CV language with vowel harmony is provided in `test.lgp`.

Configuration
-------------

Logopoeist uses a simple domain-specific programming language to describe the allowed shapes of words in its config files. The LGP language has three kinds of statements:

1. Variable declarations
2. Word Syntax rules 
3. Conditional probability rules

### Variable Declarations

Variable declarations let you give names to sets (or classes) of characters, so that you can use them multiple times in the phonotactic rules. Variable declarations have the form

    {C-var} = {C-class}

where `{C-var}` is a character class variable name, and `{C-class}` is either another variable name, or a literal character class. Character class variables always begin with a hash symbol (`#`). Literal character classes have the form

    <{char} *{Frequency} ...>

where `{char}` is some string of characters representing a phoneme (not limiting it to a single typable character allows you to treat digraphs, trigraphs, and other sequences as single characters from the point of view of phonotactics), and `{Frequency}` is a number specifying the relative frequency of that phoneme compared to others in the same set. The `*{Frequency}` setting after each phoneme is optional, and will be automatically set to 1 if not specified.

### Word Syntax Rules

Word Syntax rules describe the high-level phonotactic structure of a language in terms of a probabilistic context-free grammar.

Syntax rules have the general form

    {S-var} -> [Replacement List] *{Frequency}

Where `{S-var}` is a syntax variable, `[Replacement List]` is a space-separated list of syntax variables, character class variables, or literal character classes. As with character classes, the frequency specification is optional, and defaults to 1 if not specified; in this case, it specifies how frequently a particular substitution rule will be applied when more than one is available for the same syntax variable. Syntax variables always begin with a dollar symbol (`$`).

The left-hand symbol for the _first_ syntax rule in a configuration file will be used as the starting symbol for the probabilistic grammar.

### Conditional Probability Rules

Conditional Probability rules are used to manipulate the frequency of certain phonemes in particular contexts, given by the two preceding phonemes in a given word. They have the general form

	_ {C-class} ... -> {C-class}

where each `{C-class}` is either a character class variable or a character class literal. The leading underscore is optional; if present, it indicates a word boundary. Thus, `_ -> #A` specifies a distribution of phonemes that can come at the beginning of a word, while `_ #A -> #B` specifies a distribution of phonemes that can come second in a word.

Classes preceding the arrow (`->`) are known as _conditioning classes_. The frequencies of the conditioning classes are ignored, and they are used to generate _conditioning n-grams_ from all combinations of phonemes that can be selected from each class in order (i.e., the Cartesian product). The conditional distribution to the right of the arrow is assigned to all positions following any conditioning n-gram derived from the current rule.

A variant on conditional probability rules can be used to indicate that certain phonemes are disallowed in specific position. These have the form

	_ {C-class} ... !> {C-class}

Note the differently shaped arrow- `!>` instead of `->`. In these rules, frequencies are ignored in both conditioning and conditional classes; members of the conditional class are assigned zero probability after any conditioning n-grams derived from these rules, overriding whatever other distributions they might have had due to other rules.

Word Generation
---------------

When generating a word, Logopoeist first creates a phonotactic template for that word using the probabilistic word grammar by replacing syntax variables until a list of only character classes (indicated either by variables or literal character classes) is left.

At that point, the template is filled in from left-to-right by

1. Examining the distribution for the character class.
2. Intersecting the template distribution with any conditional distributions given by the previous phonemes already generated for the word.
3. Randomly selecting a character from the resulting distribution.

If there are no conditional distribution rules that apply at a certain position, the conditional distribution is implicitly a uniform distribution over all phonemes in the language. Intersecting distributions means eliminating phonemes that are not present in both distributions, and the multiplying the relative frequencies for each phoneme from each distribution.

TODO
----

There are plenty of ways that Logopoeist could be improved, so feel free to make suggestions and/or pull requests!