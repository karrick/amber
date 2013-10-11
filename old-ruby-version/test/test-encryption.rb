#!/usr/bin/env ruby

require 'fileutils'
require 'test/unit'

$:.unshift File.expand_path(File.join(File.dirname(__FILE__), '..', 'bin'))
load 'amber'

# NOTE: would be interesting to have unit tests that load amber and
# invoke its methods, while also having integration tests that merely
# execute the binary and check for expected behavior

class TestEncryption < Test::Unit::TestCase

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
  # file_hash

  def test_sha1_hash
    testfile = 'foo'
    File.open(testfile, 'wb') do |io|
      1.upto(100) {|x| io.puts x}
    end
    assert_equal("8084f0f10255c5e26605a1cb1f51c5e53f92df40",
                 file_hash(testfile))

    testfile = 'bar'
    File.open(testfile, 'wb') do |io|
      ['a'..'z'].each {|x| io.puts x}
    end
    assert_equal("2a092d186e21043727b874c85dc92723896ff6ab",
                 file_hash(testfile))
  end

  def test_ensure_file_hash_correct_does_not_raise_when_correct
    with_temp_file_contents("foo\nbar\nbaz\n") do |temp|
      hash = file_hash(temp)
      ensure_file_hash_correct(temp, hash)
    end
  end

  def test_ensure_file_hash_correct_raises_when_not_correct
    assert_raises RuntimeError do
      with_temp_file_contents("foo\nbar\nbaz\n") do |temp|
        ensure_file_hash_correct(temp, 'this constant will not match the file hash')
      end
    end
  end

  ################
  # encrypt_file / decrypt_file

  def test_null_encryption_and_decryption
    $encryption = nil
    key = 'foobarbaz'
    with_temp_file_contents('null encryption test') do |original|
      original_hash = file_hash(original)
      Shex.with_temp do |encrypted_temp|
        encrypt_file(original, encrypted_temp, key)

        assert_equal(original_hash, file_hash(encrypted_temp),
                     "null encrypted file should be identical to original file")

        Shex.with_temp do |decrypted_temp|
          decrypt_file(encrypted_temp, decrypted_temp, key)
          assert_equal(original_hash, file_hash(decrypted_temp),
                       "decrypted file should match the original")
        end
      end
    end
  end

  def test_aes_encryption_and_decryption
    $encryption = :aes
    key = 'foobarbaz'
    with_temp_file_contents('foo') do |original|
      original_hash = file_hash(original)
      Shex.with_temp do |encrypted_temp|
        encrypt_file(original, encrypted_temp, key)

        assert_not_equal(original_hash, file_hash(encrypted_temp),
                         "encrypted file should not be identical to original file")

        Shex.with_temp do |decrypted_temp|
          decrypt_file(encrypted_temp, decrypted_temp, key)
          assert_equal(original_hash, file_hash(decrypted_temp),
                       "decrypted file should match the original")
        end
      end
    end
  end

end
