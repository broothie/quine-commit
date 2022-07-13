require 'async'
require 'fileutils'

NUM_WORKERS = ENV.fetch('NUM_WORKERS', 1)
CHARACTERS = (('0'..'9').to_a + ('a'..'f').to_a).freeze

start_time = Time.now

Async do |task|
  NUM_WORKERS.times do |worker|
    task.async do
      repo_path = File.join('clones', start_time.to_i.to_s, "#{worker}-self-referential-commit")
      FileUtils.mkdir_p(repo_path)

      puts `git clone https://github.com/broothie/self-referential-commit.git #{repo_path}`

      loop do
        short_sha = Array.new(7) { CHARACTERS.sample }.join
        puts "attempt with short sha #{short_sha}"

        message = "short sha: #{short_sha}"
        puts output = `git -C #{repo_path} commit --allow-empty -m '#{message}'`.chomp

        if output == "[main #{short_sha}] #{message}"
          File.write('short.sha', short_sha)
          puts "success! the lucky short sha is #{short_sha}, under #{repo_path}"
          task.stop
        else
          puts `git -C #{repo_path} reset --hard HEAD~`
          puts `git -C #{repo_path} gc`
        end
      end
    end
  end
end
