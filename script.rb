
CHARACTERS = ('0'..'9').to_a + ('a'..'f').to_a

def random_hex_char
  CHARACTERS.sample
end

def random_short_sha
  Array.new(7) { random_hex_char }.join
end

if __FILE__ == $0
  counter = 0
  loop do
    attempt = counter + 1
    short_sha = random_short_sha
    puts "attempt #{attempt}, short sha is #{short_sha}" if (counter % 1000).zero?

    message = "attempt #{attempt}: #{short_sha}"
    output = `git commit --allow-empty -m '#{message}'`.chomp

    if output == "[main #{short_sha}] #{message}"
      puts "success! the lucky short sha is #{short_sha}"
      break
    else
      `git reset --hard HEAD~`
    end

    counter += 1
  end
end
