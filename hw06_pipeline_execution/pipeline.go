package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func drain(in In) {
	for range in {
	}
}

func doneWrapper(done In, in In) Out {
	out := make(Bi)

	go func() {
		defer func() {
			close(out)
			drain(in)
		}()

		for {
			select {
			case <-done:
				return
			default:
			}

			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}

				out <- v
			}
		}
	}()

	return out
}

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if in == nil {
		return nil
	}

	out := doneWrapper(done, in)

	for _, stage := range stages {
		if stage == nil {
			continue
		}

		out = doneWrapper(done, stage(out))
	}

	return out
}
