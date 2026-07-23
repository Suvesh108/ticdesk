const { spawn } = require('child_process');

function runCommand(command, args, options = {}) {
  const child = spawn(command, args, {
    stdio: 'inherit',
    shell: true,
    ...options,
  });

  child.on('exit', (code) => {
    if (code !== 0) {
      process.exit(code);
    }
  });

  return child;
}

const processes = [
  { command: 'go', args: ['run', './cmd/server'] },
];

processes.forEach((proc) => runCommand(proc.command, proc.args));
