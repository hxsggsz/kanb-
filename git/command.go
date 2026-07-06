package git

import "context"

type Command[T any] interface {
	Args() []string
	Parse(string) (T, error)
}

func Execute[T any](ctx context.Context, runner Runner, cmd Command[T]) (T, error) {
	out, err := runner.Run(ctx, cmd.Args()...)
	if err != nil {
		var zero T
		return zero, err
	}
	return cmd.Parse(out)
}
