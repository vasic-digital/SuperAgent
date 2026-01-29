#!/usr/bin/env ruby
# Universal HTTP service wrapper for Ruby formatters (rubocop, standardrb)

require 'webrick'
require 'json'
require 'tempfile'
require 'optparse'

class FormatterServiceHandler < WEBrick::HTTPServlet::AbstractServlet
  def initialize(server, formatter_name, formatter_binary)
    super(server)
    @formatter_name = formatter_name
    @formatter_binary = formatter_binary
  end

  def do_POST(request, response)
    if request.path != '/format'
      response.status = 404
      response['Content-Type'] = 'application/json'
      response.body = { success: false, error: 'Not Found' }.to_json
      return
    end

    begin
      data = JSON.parse(request.body)
      content = data['content'] || ''
      options = data['options'] || {}

      if content.empty?
        response.status = 400
        response['Content-Type'] = 'application/json'
        response.body = { success: false, error: 'content field is required' }.to_json
        return
      end

      # Format the code
      result = format_code(content, options)

      response.status = 200
      response['Content-Type'] = 'application/json'
      response.body = result.to_json

    rescue JSON::ParserError
      response.status = 400
      response['Content-Type'] = 'application/json'
      response.body = { success: false, error: 'Invalid JSON' }.to_json
    rescue => e
      response.status = 500
      response['Content-Type'] = 'application/json'
      response.body = { success: false, error: e.message }.to_json
    end
  end

  def do_GET(request, response)
    if request.path == '/health'
      begin
        version = `#{@formatter_binary} --version 2>&1`.strip

        response.status = 200
        response['Content-Type'] = 'application/json'
        response.body = {
          status: 'healthy',
          formatter: @formatter_name,
          version: version
        }.to_json
      rescue => e
        response.status = 503
        response['Content-Type'] = 'application/json'
        response.body = { status: 'unhealthy', error: e.message }.to_json
      end
    else
      response.status = 404
      response['Content-Type'] = 'application/json'
      response.body = { success: false, error: 'Not Found' }.to_json
    end
  end

  private

  def format_code(content, options)
    begin
      # Create temp file for rubocop (requires file input)
      Tempfile.create(['code', '.rb']) do |f|
        f.write(content)
        f.flush

        # Execute rubocop with auto-correct
        cmd = "#{@formatter_binary} --auto-correct #{f.path} 2>&1"
        output = `#{cmd}`
        exit_code = $?.exitstatus

        # Read formatted content
        formatted = File.read(f.path)

        return {
          success: true,
          content: formatted,
          changed: content != formatted,
          formatter: @formatter_name
        }
      end
    rescue => e
      return {
        success: false,
        error: e.message,
        formatter: @formatter_name
      }
    end
  end
end

def main
  options = {
    formatter: nil,
    port: nil,
    host: '0.0.0.0',
    binary: nil
  }

  OptionParser.new do |opts|
    opts.banner = 'Usage: formatter-service.rb [options]'

    opts.on('--formatter FORMATTER', 'Formatter name (rubocop, standardrb)') do |f|
      options[:formatter] = f
    end

    opts.on('--port PORT', Integer, 'HTTP port') do |p|
      options[:port] = p
    end

    opts.on('--host HOST', 'HTTP host (default: 0.0.0.0)') do |h|
      options[:host] = h
    end

    opts.on('--binary BINARY', 'Formatter binary path') do |b|
      options[:binary] = b
    end
  end.parse!

  unless options[:formatter] && options[:port]
    puts 'Error: --formatter and --port are required'
    exit 1
  end

  options[:binary] ||= options[:formatter]

  # Create HTTP server
  server = WEBrick::HTTPServer.new(
    Port: options[:port],
    BindAddress: options[:host],
    Logger: WEBrick::Log.new($stderr, WEBrick::Log::INFO),
    AccessLog: []
  )

  # Mount handler
  server.mount('/format', FormatterServiceHandler, options[:formatter], options[:binary])
  server.mount('/health', FormatterServiceHandler, options[:formatter], options[:binary])

  puts "🚀 #{options[:formatter]} formatter service started on #{options[:host]}:#{options[:port]}"
  puts "   Health: http://#{options[:host]}:#{options[:port]}/health"
  puts "   Format: POST http://#{options[:host]}:#{options[:port]}/format"

  trap('INT') { server.shutdown }

  server.start
end

main if __FILE__ == $PROGRAM_NAME
