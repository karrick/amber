#!/usr/bin/env ruby

require 'fileutils'
require 'test/unit'

$:.unshift File.expand_path(File.join(File.dirname(__FILE__), '..', 'bin'))
load 'amber'

# NOTE: would be interesting to have unit tests that load amber and
# invoke its methods, while also having integration tests that merely
# execute the binary and check for expected behavior

class TestFileManagement < Test::Unit::TestCase

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

  ################

  def test_with_temp_file_contents_catches_missing_block
    assert_raises ArgumentError do
      with_temp_file_contents('foo')
    end
  end

  def test_with_temp_file_contents_writes_file
    tempfile = nil
    with_temp_file_contents('foo') do |temp|
      tempfile = temp
      assert_equal('foo', File.read(temp))
    end

    assert(!File.file?(tempfile))
  end

  ################

  def test_create_directory_and_install_file
    with_temp_file_contents("foo\nbar\nbaz") do |temp|
      FileUtils.rm_rf("foo") if File.exists?("foo")
      create_directory_and_install_file(temp, "foo/bar/baz")
      assert_equal('01e06a68df2f0598042449c4088842bb4e92ca75', 
                   file_hash("foo/bar/baz"))
    end
  end

  ################

  def test_address_to_pathname_puts_files_in_amber_archive
    address = '01e06a68df2f0598042449c4088842bb4e92ca75'
    assert_match(/^.amber\/archive\//, address_to_pathname(address))
  end

  def test_address_to_pathname_puts_splits_hash
    address = '01e06a68df2f0598042449c4088842bb4e92ca75'
    result = address_to_pathname(address)
    result.gsub!(/^.amber\/archive\//, '')
    assert_match(/\//, result)
    assert_equal(address, result.gsub('/', ''))
  end

end
