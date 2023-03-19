package prods

// Authors contains Defacto2 people rolls.
type Authors struct {
	Text  []string // credit_text, writer
	Code  []string // credit_program, programmer/coder
	Art   []string // credit_illustration, artist/graphics
	Audio []string // credit_audio, musician/sound
}

// Authors parses Demozoo authors and reclassifies them into Defacto2 people rolls.
func (p *ProductionsAPIv1) Authors() Authors {
	a := Authors{}
	for _, c := range p.Credits {
		if c.Nick.Releaser.IsGroup {
			continue
		}
		switch category(c.Category) {
		case Text:
			a.Text = append(a.Text, c.Nick.Name)
		case Code:
			a.Code = append(a.Code, c.Nick.Name)
		case Graphics:
			a.Art = append(a.Art, c.Nick.Name)
		case Music:
			a.Audio = append(a.Audio, c.Nick.Name)
		case Magazine:
			// do nothing.
		}
	}
	return a
}
