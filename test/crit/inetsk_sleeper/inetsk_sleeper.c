#define _GNU_SOURCE
#include <arpa/inet.h>
#include <netinet/in.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <unistd.h>

/*
 * Bind a TCP listen socket and sleep in a daemonized child. The parent
 * prints the child PID and exits, and the child closes its stdio, so that
 * $(shell ...) in the Makefile gets EOF and returns immediately instead of
 * blocking on the sleeping process. Used to produce files.img entries with
 * INETSK for crit humanize integration tests.
 */
int main(void)
{
	pid_t pid;
	int res = EXIT_FAILURE;
	int start_pipe[2];

	if (pipe(start_pipe)) {
		perror("pipe");
		return 1;
	}

	pid = fork();
	if (pid < 0) {
		perror("fork");
		return 1;
	}

	if (pid == 0) {
		int fd;
		struct sockaddr_in addr = { 0 };

		close(start_pipe[0]);

		if (setsid() < 0) {
			perror("setsid");
			goto child_out;
		}

		fd = socket(AF_INET, SOCK_STREAM, 0);
		if (fd < 0) {
			perror("socket");
			goto child_out;
		}

		addr.sin_family = AF_INET;
		addr.sin_addr.s_addr = htonl(INADDR_ANY);
		addr.sin_port = htons(0);

		if (bind(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
			perror("bind");
			goto child_out;
		}
		if (listen(fd, 1) < 0) {
			perror("listen");
			goto child_out;
		}

		close(STDIN_FILENO);
		close(STDOUT_FILENO);
		close(STDERR_FILENO);

		res = EXIT_SUCCESS;
		if (write(start_pipe[1], &res, sizeof(res)) != sizeof(res))
			_exit(1);
		close(start_pipe[1]);

		while (1)
			sleep(3600);

	child_out:
		if (write(start_pipe[1], &res, sizeof(res)) != sizeof(res))
			_exit(1);
		close(start_pipe[1]);
		_exit(1);
	}

	close(start_pipe[1]);
	if (read(start_pipe[0], &res, sizeof(res)) != sizeof(res))
		res = EXIT_FAILURE;
	close(start_pipe[0]);

	if (res == EXIT_SUCCESS)
		printf("%d\n", pid);

	return res;
}
