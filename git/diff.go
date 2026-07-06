package git

type DiffArgs struct {
	Show bool
	Args []string
}

type DiffCommand struct {
	DiffArgs
	Parser  *UnifiedParser
	Aligner LineAligner
}

func (c *DiffCommand) Args() []string {
	if c.DiffArgs.Show {
		return append([]string{"show", "--no-color", "--unified=3"}, c.DiffArgs.Args...)
	}
	return append([]string{"diff", "--no-color", "--unified=3"}, c.DiffArgs.Args...)
}

func (c *DiffCommand) Parse(raw string) ([]SideBySideDiff, error) {
	files, err := c.Parser.Parse(raw)
	if err != nil {
		return nil, err
	}
	result := make([]SideBySideDiff, len(files))
	for i, f := range files {
		result[i] = SideBySideDiff{
			OldPath: f.OldPath,
			NewPath: f.NewPath,
			Status:  f.Status,
			Hunks:   c.Aligner.Align(f.Hunks),
		}
	}
	return result, nil
}
