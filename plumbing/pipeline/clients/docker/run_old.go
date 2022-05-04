package docker

// // PipelineVolumePath refers to the path where the compiled pipeline is mounted in the container
// const PipelineVolumePath = "/var/pipeline"
//
// type RunOpts struct {
// 	Image   string
// 	Command string
// 	Volumes []string
// 	Args    []string
// 	Env     []string
//
// 	Stdout io.Writer
// 	Stderr io.Writer
// }
//
// func (opts RunOpts) WithPipelinePath(path string) RunOpts {
// 	if opts.Volumes == nil {
// 		opts.Volumes = []string{}
// 	}
//
// 	opts.Volumes = append(opts.Volumes, fmt.Sprintf("%s:%s", path, "/var/pipeline"))
//
// 	return opts
// }
//
// func RunArgs(opts RunOpts) []string {
// 	var (
// 		volumes = []string{}
// 		env     = []string{}
// 	)
//
// 	for _, v := range opts.Volumes {
// 		volumes = append(volumes, "-v", v)
// 	}
//
// 	for _, v := range opts.Env {
// 		env = append(env, "-e", v)
// 	}
//
// 	args := []string{"run", "--rm"}
// 	args = append(args, volumes...)
// 	args = append(args, env...)
// 	args = append(args, opts.Image)
// 	args = append(args, opts.Command)
// 	args = append(args, opts.Args...)
//
// 	return args
// }
//
// func Run(ctx context.Context, opts RunOpts) error {
// 	cmd := exec.CommandContext(ctx, "docker", RunArgs(opts)...)
// 	cmd.Stdout = opts.Stdout
// 	cmd.Stderr = opts.Stderr
//
// 	return cmd.Run()
// }
