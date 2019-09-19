package compiler

//Language is a written langauge.
type Language int

//This is a list of available languages.
const (
	English Language = iota

	Afrikaans
	Albanian
	Amharic
	Arabic
	Armenian
	Azerbaijani
	Basque
	Belarusian
	Bengali
	Bosnian
	Bulgarian
	Burmese
	Catalan
	Cebuano
	Chichewa
	Chinese
	Corsican
	Croatian
	Czech
	Danish
	Dutch
	Esperanto
	Estonian
	Filipino
	Finnish
	French
	Frisian
	Galician
	Georgian
	German
	Greek
	Gujarati
	HaitianCreole
	Hausa
	Hawaiian
	Hebrew
	Hindi
	Hmong
	Hungarian
	Icelandic
	Igbo
	Indonesian
	Irish
	Italian
	Japanese
	Javanese
	Kannada
	Kazakh
	Khmer
	Korean
	Kurdish
	Kyrgyz
	Lao
	Latin
	Latvian
	Lithuanian
	Luxembourgish
	Macedonian
	Malagasy
	Malay
	Malayalam
	Maltese
	Maori
	Marathi
	Mongolian
	Nepali
	Norwegian
	Pashto
	Persian
	Polish
	Portuguese
	Punjabi
	Romanian
	Russian
	Samoan
	ScotsGaelic
	Serbian
	Sesotho
	Shona
	Sindhi
	Sinhala
	Slovak
	Slovenian
	Somali
	Spanish
	Sundanese
	Swahili
	Swedish
	Tajik
	Tamil
	Telugu
	Thai
	Turkish
	Ukrainian
	Urdu
	Uzbek
	Vietnamese
	Welsh
	Xhosa
	Yiddish
	Yoruba
	Zulu

	Klingon

	All
)

//String is a translatable string.
type String map[Language]string
