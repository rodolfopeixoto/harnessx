ENV['RACK_ENV'] = 'test'

require 'rspec'
require 'rack/test'
require_relative '../app'

RSpec.configure do |c|
  c.include Rack::Test::Methods
  c.expect_with(:rspec) { |e| e.syntax = :expect }
end

def app
  Sinatra::Application
end
