#!/usr/bin/env ruby

require 'fileutils'
require 'test/unit'

$:.unshift File.expand_path(File.join(File.dirname(__FILE__), '..', 'bin'))
load 'amber'

# NOTE: would be interesting to have unit tests that load amber and
# invoke its methods, while also having integration tests that merely
# execute the binary and check for expected behavior

class TestInitialization < Test::Unit::TestCase

  def setup
    $save_dir = Dir.pwd

    $test_root = File.expand_path(File.join(File.dirname(__FILE__),
                                            'data',
                                            File.basename(__FILE__, '.*')))
    FileUtils.rm_rf($test_root)
    FileUtils.mkdir_p(File.join($test_root, '.amber'))
    Dir.chdir($test_root)
  end

  def teardown
    Dir.chdir $save_dir
    FileUtils.rm_rf($test_root)
  end

  def test_find_repository_root_already_there
    assert_equal($test_root, repository_root)
  end

  def test_find_repository_root_inside
    FileUtils.mkdir_p('foo/bar')
    Dir.chdir('foo/bar')
    assert_equal($test_root, repository_root)
  end

  def test_find_repository_root_cannot_find
    # test requires parent of HOME directory to not have .amber
    # directory
    assert_raises RuntimeError do
      Dir.chdir(File.join(ENV['HOME'], '..'))
      repository_root
    end
  end
  
  ################

  def test_amber_ignores_returns_empty_when_not_found
    ignores_filename = File.join(repository_root, '.amber-ignore')
    FileUtils.rm(ignores_filename) if File.file?(ignores_filename)
    assert_equal([], amber_ignores(repository_root))
  end

  def test_amber_ignores_returns_empty_when_not_found
    ignores_filename = File.join(repository_root, '.amber-ignore')
    File.open(ignores_filename, 'w') do |io|
      ["foo", "bar*", "*baz"].each do |line|
        io.puts line
      end
    end
    assert_equal(['foo', 'bar*', '*baz'], 
                 amber_ignores(repository_root))
  end

end
