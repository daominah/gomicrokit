package textproc

import "testing"

var paragraphs = []string{
	`I can't say how every time I ever put my arms around you I felt that I was home,`,
	`The scariest moment in my time with Team Secret was during our practices,
when Puppey would walk around with a machete and talk about how he always wanted
to see what the inside of a human looked like. He said he had experimented on
animals before and he wanted to go for the real thing.`,
	`Sodium, atomic number 11, was first isolated by Peter Dager in 1807.
	A chemical component of salt, he named it Na in honor of the saltiest region
on earth, North America.`,
	` Something stupid happened

I'm not allowed to have my own cell phone so my dad forced me to use his phone number.
My dad has a steam too and uses the same number. today my brother used my dads
account and cheated and now my main account is VAC banned. It's true and here is proof,
my father will now write too: Hello I'm the father and what my son says is true,
he did not cheat, it was his brother on my account. Please unban him valve

sincerely the father

Pls unban `,
}

func BenchmarkTextToSyls(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for _, para := range paragraphs {
			_ = TextToWords(para)
		}
	}
}

func BenchmarkTextToNGrams(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for _, para := range paragraphs {
			nGrams := TextToNGrams(para, 2)
			_ = nGrams
		}
	}
}
