#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <sys/mman.h>
#include <fcntl.h>
#include <unistd.h>
#include <sched.h>
#include <signal.h>
#include <string.h>

#define STKS 4096

#ifndef CLONE_NEWPID
#define CLONE_NEWPID 0x20000000
#endif

static int do_test(void *logf)
{
	int fd, i = 0;

	// Detach from the parent terminal
	setsid();

	// Close standard file descriptors
	close(STDIN_FILENO);
	close(STDOUT_FILENO);
	close(STDERR_FILENO);

	// Open /dev/null as stdin
	fd = open("/dev/null", O_RDONLY);
	if (fd != STDIN_FILENO) {
		dup2(fd, STDIN_FILENO);
		close(fd);
	}

	// Open the log file and redirect stdout and stderr to it
	fd = open(logf, O_WRONLY | O_TRUNC | O_CREAT, 0600);
	dup2(fd, STDOUT_FILENO);
	dup2(fd, STDERR_FILENO);
	if (fd != STDOUT_FILENO && fd != STDERR_FILENO)
		close(fd);

	while (1) {
		printf("%d\n", i++);
		fflush(stdout);
		sleep(1);
	}

	return 0;
}

int main(int argc, char *argv[])
{
	int pid;
	size_t size;
	void *stk;
	int flags = 0;
	int fd = 0;

	if (argc < 4) {
		fprintf(stderr, "Usage: %s <size> <mapping_flags (s|p[a])> <log_file>\n", argv[0]);
		fprintf(stderr, "Example: %s 4096 sp /tmp/mmapper.log\n", argv[0]);
		return 1;
	}

	size = atoi(argv[1]);
	if (size <= 0) {
		fprintf(stderr, "Invalid input size\n");
		return 1;
	}

	char *mapping_flags = argv[2];

	// Parse mapping flags argument
	for (int i = 0; i < strlen(mapping_flags); i++) {
		switch (mapping_flags[i]) {
		case 's':
			flags |= MAP_SHARED;
			break;
		case 'p':
			flags |= MAP_PRIVATE;
			break;
		case 'a':
			flags |= MAP_ANONYMOUS;
			break;
		}
	}

	// If using MAP_PRIVATE or MAP_SHARED without MAP_ANONYMOUS, create a temporary file
	if (flags == MAP_SHARED || flags == MAP_PRIVATE) {
		char filename[] = "/tmp/crit-mmapper.XXXXXX";
		fd = mkstemp(filename);
		if (fd == -1) {
			perror("mkstemp() failed");
			return 1;
		}
		ftruncate(fd, size);

		// Clean-up the temporary file
		unlink(filename);
	}

	// Create the memory mapping
	stk = mmap(NULL, size, PROT_READ | PROT_WRITE, flags, fd, 0);
	if (stk == MAP_FAILED) {
		perror("mmap() failed");
		return 1;
	}

	// Create a child process
	pid = clone(do_test, stk + STKS, SIGCHLD | CLONE_NEWPID, argv[3]);
	if (pid < 0) {
		perror("clone() failed");
		return 1;
	}
	printf("%d\n", pid);

	return 0;
}
