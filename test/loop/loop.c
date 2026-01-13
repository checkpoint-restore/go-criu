#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdlib.h>
#include <stdio.h>

int main(void)
{
	pid_t pid;
	pid_t sid;
	int res = EXIT_FAILURE;
	int start_pipe[2];

	// Set a PASSWORD environment variable to test the search pattern
	// within process memory pages using regex metacharacters.
	if (setenv("PASSWORD", "123 Hello.*?[^]@WORLD(|x)", 1) != 0) {
		perror("setenv");
		return 1;
	}

	if (pipe(start_pipe)) {
		perror("pipe failed!");
		goto out;
	}

	pid = fork();
	if (pid < 0) {
		perror("fork failed!");
		goto out;
	}

	if (pid == 0) {
		close(start_pipe[0]);

		sid = setsid();
		if (sid < 0) {
			perror("setsid failed!");
			res = EXIT_FAILURE;
			write(start_pipe[1], &res, sizeof(res));
			close(start_pipe[1]);
			exit(1);
		}

		// Create a file descriptor for "crit x ./ fd" test
		open("/dev/null", O_RDONLY);

		chdir("/");
		close(STDIN_FILENO);
		close(STDOUT_FILENO);
		close(STDERR_FILENO);

		res = EXIT_SUCCESS;
		write(start_pipe[1], &res, sizeof(res));
		close(start_pipe[1]);

		/* Create a process with task_state="dead" */
		if (fork() == 0)
			_exit(0);

		while (1) {
			sleep(1);
		}
	}

	close(start_pipe[1]);
	read(start_pipe[0], &res, sizeof(res));
	close(start_pipe[0]);

out:
	if (res == EXIT_SUCCESS)
		printf("%d\n", pid);
	return res;
}
