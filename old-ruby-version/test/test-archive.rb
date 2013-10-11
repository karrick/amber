#!/usr/bin/env ruby

require 'fileutils'
require 'test/unit'

$:.unshift File.expand_path(File.join(File.dirname(__FILE__), '..', 'bin'))
load 'amber'

# NOTE: would be interesting to have unit tests that load amber and
# invoke its methods, while also having integration tests that merely
# execute the binary and check for expected behavior

class TestArchive < Test::Unit::TestCase

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
    Dir.chdir($save_dir)
    FileUtils.rm_rf($test_root)
  end

  ################
  # encrypt_file_with_hash

  def test_should_use_hash_of_original_as_encryption_key
    with_temp_file_contents("One Mississippi\nTwo Mississippi\n") do |original|
      original_hash = file_hash(original)

      Shex.with_temp do |encrypted_temp|
        item = encrypt_file_with_hash(original, encrypted_temp)

        assert_equal(original_hash, item['key'],
                     "key value should be hash of plain text")

        Shex.with_temp do |decrypted_temp|
          decrypt_file(encrypted_temp, decrypted_temp, 
                       original_hash)
          assert_equal(original_hash, file_hash(decrypted_temp),
                       "should have used hash of plain text as encryption key")
        end
      end
    end
  end

  ################
  # store_file_in_archive

  def test_store_file_in_archive_should_use_hash_for_address
    with_temp_file_contents("Three blind mice") do |temp|
      Shex.with_temp do |etemp|
        encrypt_file_with_hash(temp, etemp)

        item = store_file_in_archive(etemp)

        ehash = file_hash(etemp)
        assert_equal(ehash, item['address'])
        assert_equal(ehash, file_hash(address_to_pathname(item['address'])))
      end
    end
  end

  ################
  # verify_hash_of_file_in_archive

  def test_verify_hash_of_file_in_archive_true
    with_temp_file_contents("The quick brown fox...") do |temp|
      item = store_file_in_archive(temp)
      assert(verify_hash_of_file_in_archive(item['address']),
             "should detect when hash of archived file is correct")
    end
  end

  def test_verify_hash_of_file_in_archive_false
    with_temp_file_contents("E pluribus unim...") do |temp|
      item = store_file_in_archive(temp)
      File.open(address_to_pathname(item['address']), 'a') {|io| io.write("a")}
      assert(!verify_hash_of_file_in_archive(item['address']),
             "should detect when hash of archived file is wrong")
    end
  end

  ################
  # encrypt_and_archive_file

  def test_encrypt_and_archive_file
    with_temp_file_contents("Friend, Romans, Countrymen...") do |temp|
      item = encrypt_and_archive_file(temp)
      assert_equal(File.basename(temp), item['name'])
      assert_equal('file', item['type'],
                   "type should specify type of file system object")
      assert_equal(file_hash(temp), item['key'],
                   "should store hash of original in 'key'")
      assert_match(/^[a-fA-F0-9]+$/, item['address'],
                   "address should be hexidecimal")
      assert(File.file?(address_to_pathname(item['address'])),
             "should store file in archive")
      assert_equal(item['address'], file_hash(address_to_pathname(item['address'])),
                   "hash of archived file should match address")
    end
  end

  ################
  # encrypt_and_archive_symlink

  def test_encrypt_and_archive_symlink
    File.open('target', 'wb') {|io| io.puts 'target'}
    File.symlink('target', 'symlink')
    assert(File.symlink?('symlink'),
           "test framework should be able to create a symlink")
    assert_equal({ 'type' => 'symlink',
                   'referent' => 'target',
                   'name' => 'symlink',
                   'mode' => File.lstat('symlink').mode},
                 encrypt_and_archive_symlink('symlink'),
                 "archiving symlink should return special hash")
  end

  ################
  # directory_children

  def test_directory_children_includes_dot_files_and_ignores_ignored_files
    File.open('.amber-ignore', 'wb') do |io|
      io.puts 'IGNORE*'
      io.puts '*-*'
    end

    frazzle = 'frazzle'
    frazzle_dazzle = File.join(frazzle, 'dazzle')
    frazzle_foo = File.join(frazzle, 'foo')
    frazzle_dazzle_bar = File.join(frazzle_dazzle, '.bar')
    frazzle_baz = File.join(frazzle, 'baz')
    frazzle_dazzle_faz = File.join(frazzle_dazzle, 'faz')
    ignore = File.join(frazzle, 'IGNORE ME')
    this_too = File.join(frazzle_dazzle, 'this-too')

    FileUtils.mkdir_p(frazzle_dazzle)
    File.open(frazzle_foo,'wb') {|io| io.puts 'frazzle_foo'}
    File.open(frazzle_dazzle_bar,'wb') {|io| io.puts 'frazzle_dazzle_bar'}
    File.open(ignore, 'wb') {|io| io.puts ignore }
    File.open(this_too, 'wb') {|io| io.puts this_too }
    File.symlink(frazzle_dazzle_bar, frazzle_baz)
    File.symlink(frazzle_foo, frazzle_dazzle_faz)

    assert_equal(['frazzle/baz',
                  'frazzle/dazzle',
                  'frazzle/foo'],
                 directory_children(frazzle).sort)
  end

  ################
  # encrypt_and_archive_directory

  def test_encrypt_and_archive_directory
    directory = 'frazzle'
    FileUtils.mkdir_p(directory)
    ['a','b','.c'].each do |letter|
      File.open(File.join(directory, letter), 'wb') do |io|
        io.puts letter
      end
    end
    mode = File.stat(directory).mode
    item = encrypt_and_archive_directory(directory)
    assert_equal('directory', item['type'])
    assert_equal(File.basename(directory), item['name'])
    assert_equal(File.stat(directory).mode, item['mode'])
    assert_match(/^[a-fA-F0-9]+$/, item['address'])
    assert_match(/^[a-fA-F0-9]+$/, item['key'])
  end

  ################
  # encrypt_and_archive

  def test_encrypt_and_archive_raises_if_not_supported_file_type
    # NOTE: The following tests were written to work on Debian: they
    # may need to be modified to work on more systems.

    # blockdev?
    assert_raises RuntimeError do
      encrypt_and_archive(Dir.glob('/dev/loop0').first)
    end
    # chardev?
    assert_raises RuntimeError do
      encrypt_and_archive(Dir.glob('/dev/zero').first)
    end

    # pipe?
    # socket?
  end

  def test_encrypt_and_archive_works_with_files
    with_temp_file_contents("Does it work with files?") do |temp|
      item = encrypt_and_archive(temp)
      assert_equal('file', item['type'])
      assert_equal(File.basename(temp), item['name'])
      assert_equal(File.stat(temp).mode, item['mode'])
      assert_match(/^[a-fA-F0-9]+$/, item['address'])
      assert_match(/^[a-fA-F0-9]+$/, item['key'])
    end
  end

  def test_encrypt_and_archive_works_with_symbolic_links
    with_temp_file_contents("Does it work with symbolic links?") do |temp|
      File.symlink(temp, 'some-link')
      item = encrypt_and_archive('some-link')
      assert_equal('symlink', item['type'])
      assert_equal('some-link', item['name'])
      assert(item['location'].nil?)
      assert(item['hash'].nil?)
      assert_equal(File.lstat('some-link').mode, item['mode'])
    end
  end

  def test_encrypt_and_archive_recurses_with_directory
    directory = 'frazzle/dazzle'
    FileUtils.mkdir_p(directory)
    directory = 'frazzle'
    ['a','dazzle/b','.c'].each do |letter|
      File.open(File.join(directory, letter), 'wb') do |io|
        io.puts letter
      end
    end
    mode = File.stat(directory).mode
    item = encrypt_and_archive_directory(directory)
    assert_equal('directory', item['type'])
    assert_equal(File.stat(directory).mode, item['mode'])
    assert_equal(File.basename(directory), item['name'])
    assert_match(/^[a-fA-F0-9]+$/, item['address'])
    assert_match(/^[a-fA-F0-9]+$/, item['key'])
  end

  ################
  # retrieve_from_archive_file

  def test_retrieve_from_archive_file
    original = "retrieve-file-test"
    backup = original + ".backup"
    File.open(original, 'wb') do |io|
      io.puts 'retrieve-file-test'
    end
    item = encrypt_and_archive(original)
    FileUtils.mv(original, backup)
    assert(!File.exists?(original))
    assert(File.exists?(backup))

    retrieve_from_archive_file(item)

    assert(File.exists?(original))
    assert_equal(File.stat(backup).mode, File.stat(original).mode)
    assert_equal(file_hash(backup), file_hash(original))
  end

  def test_retrieve_from_archive_symbolic_link
    original = "retrieve-symlink-test"
    File.open(original, 'wb') do |io|
      io.puts "retrieve-symlink-test"
    end
    symlink = 'some-link'
    File.symlink(original, symlink)
    original_mode = File.lstat(symlink).mode
    item = encrypt_and_archive(symlink)
    FileUtils.rm_rf(symlink)

    retrieve_from_archive_symlink(item)

    assert(File.symlink?(symlink))
    assert_equal(original, File.readlink(symlink))
    assert_equal(original_mode, File.lstat(symlink).mode)
  end

  def test_read_directory_contents_from_archive
    frazzle = 'frazzle'
    frazzle_dazzle = File.join(frazzle, 'dazzle')
    frazzle_foo = File.join(frazzle, 'foo')
    frazzle_dazzle_bar = File.join(frazzle_dazzle, 'bar')
    frazzle_baz = File.join(frazzle, 'baz')
    frazzle_dazzle_faz = File.join(frazzle_dazzle, 'faz')

    FileUtils.mkdir_p(frazzle_dazzle)
    File.open(frazzle_foo,'wb') {|io| io.puts 'frazzle_foo'}
    File.open(frazzle_dazzle_bar,'wb') {|io| io.puts 'frazzle_dazzle_bar'}
    File.symlink(frazzle_dazzle_bar, frazzle_baz)
    File.symlink(frazzle_foo, frazzle_dazzle_faz)

    frazzle_mode = File.stat(frazzle).mode
    frazzle_dazzle_mode = File.stat(frazzle_dazzle).mode
    frazzle_foo_mode = File.stat(frazzle_foo).mode
    frazzle_dazzle_bar_mode = File.stat(frazzle_dazzle_bar).mode
    frazzle_foo_hash = file_hash(frazzle_foo)
    frazzle_dazzle_bar_hash = file_hash(frazzle_dazzle_bar)
    frazzle_baz_mode = File.lstat(frazzle_baz).mode
    frazzle_dazzle_faz_mode = File.lstat(frazzle_dazzle_faz).mode

    item = encrypt_and_archive(frazzle)
    FileUtils.rm_rf(frazzle)
    assert(!File.exists?(frazzle))
    contents = read_directory_contents_from_archive(item)

    assert_equal(3, contents.length)

    child = contents.find {|x| x['name'] == 'foo'}
    assert_equal('file', child['type'])
    assert_equal(frazzle_foo_mode, child['mode'])
    assert_equal(frazzle_foo_hash, child['key'])
    assert_match(/^[a-fA-F0-9]+$/, child['address'])

    child = contents.find {|x| x['name'] == 'baz'}
    assert_equal('symlink', child['type'])
    assert_equal(frazzle_baz_mode, child['mode'])
    assert(child['key'].nil?)
    assert(child['address'].nil?)

    child = contents.find {|x| x['name'] == 'dazzle'}
    assert_equal('directory', child['type'])
    assert_equal(frazzle_dazzle_mode, child['mode'])
    assert_match(/^[a-fA-F0-9]+$/, child['key'])
    assert_match(/^[a-fA-F0-9]+$/, child['address'])
  end

  def test_retrieve_from_archive_directory
    frazzle = 'frazzle'
    frazzle_baz = File.join(frazzle, 'baz')
    frazzle_dazzle = File.join(frazzle, 'dazzle')
    frazzle_dazzle_bar = File.join(frazzle_dazzle, 'bar')
    frazzle_dazzle_faz = File.join(frazzle_dazzle, 'faz')
    frazzle_foo = File.join(frazzle, 'foo')

    FileUtils.mkdir_p(frazzle_dazzle)
    File.open(frazzle_foo,'wb') {|io| io.puts 'frazzle_foo'}
    File.open(frazzle_dazzle_bar,'wb') {|io| io.puts 'frazzle_dazzle_bar'}
    File.symlink(frazzle_dazzle_bar, frazzle_baz)
    File.symlink(frazzle_foo, frazzle_dazzle_faz)

    frazzle_baz_mode = File.lstat(frazzle_baz).mode
    frazzle_dazzle_bar_hash = file_hash(frazzle_dazzle_bar)
    frazzle_dazzle_bar_mode = File.stat(frazzle_dazzle_bar).mode
    frazzle_dazzle_faz_mode = File.lstat(frazzle_dazzle_faz).mode
    frazzle_dazzle_mode = File.stat(frazzle_dazzle).mode
    frazzle_foo_hash = file_hash(frazzle_foo)
    frazzle_foo_mode = File.stat(frazzle_foo).mode
    frazzle_mode = File.stat(frazzle).mode

    item = encrypt_and_archive(frazzle)
    FileUtils.rm_rf(frazzle)
    assert(!File.exists?(frazzle))

    retrieve_from_archive_directory(item)

    # frazzle/
    assert(File.directory?(frazzle))
    assert_equal(frazzle_mode, File.stat(frazzle).mode)

    # frazzle/foo
    assert(File.file?(frazzle_foo))
    assert_equal(frazzle_foo_mode, File.stat(frazzle_foo).mode)
    assert_equal(frazzle_foo_hash, file_hash(frazzle_foo))

    # frazzle/dazzle/
    assert(File.directory?(frazzle_dazzle))
    assert_equal(frazzle_dazzle_mode, File.stat(frazzle_dazzle).mode)

    # frazzle/dazzle/bar
    assert(File.file?(frazzle_dazzle_bar))
    assert_equal(frazzle_dazzle_bar_mode, File.stat(frazzle_dazzle_bar).mode)
    assert_equal(frazzle_dazzle_bar_hash, file_hash(frazzle_dazzle_bar))
    assert(!File.exists?(File.join(frazzle, 'bar')))

    # frazzle/baz
    assert(File.symlink?(frazzle_baz))
    assert_equal(frazzle_dazzle_bar, File.readlink(frazzle_baz))
    assert_equal(frazzle_baz_mode, File.lstat(frazzle_baz).mode)

    # frazzle/dazzle/faz
    assert(File.symlink?(frazzle_dazzle_faz))
    assert_equal(frazzle_foo, File.readlink(frazzle_dazzle_faz))
    assert_equal(frazzle_dazzle_faz_mode, File.lstat(frazzle_dazzle_faz).mode)
    assert(!File.exists?(File.join(frazzle, 'faz')))
  end

  ################
  # with_found_item

  def test_with_found_item_yields_item_when_correct_dir_is_found
    frazzle = 'frazzle'
    frazzle_dazzle = File.join(frazzle, 'dazzle')
    frazzle_dazzle_foo = File.join(frazzle_dazzle, 'foo')
    frazzle_foo = File.join(frazzle, 'foo')

    FileUtils.mkdir_p(frazzle_dazzle)
    File.open(frazzle_foo,'wb') {|io| io.puts 'foo'}
    File.open(frazzle_dazzle_foo,'wb') {|io| io.puts 'dazzle_foo'}

    frazzle_mode = File.stat(frazzle).mode
    frazzle_dazzle_mode = File.stat(frazzle_dazzle).mode

    archive = encrypt_and_archive(frazzle)
    FileUtils.rm_rf(frazzle)
    assert(!File.exists?(frazzle))
    
    with_found_item(frazzle, archive) do |item|
      assert_equal({ 'type' => 'directory',
                     'name' => File.basename(frazzle),
                     'mode' => frazzle_mode,
                     'key'  => '91a78183a8b9f5c0627a96853a26c61a22c3e5b0',
                     'address' => '8faef74f21c9a06779d658492a2297cde3e5b1fa',
                   }, item)
    end
    
    with_found_item(frazzle_dazzle, archive) do |item|
      assert_equal({
                     'type' => 'directory',
                     'name' => File.basename(frazzle_dazzle),
                     'mode' => frazzle_dazzle_mode,
                     'key'  => '61169ae49f06aa7b6c7468e15846c1b81093e226',
                     'address' => 'fd59791fa522fc1fb03cae825296d4b474d3236d',
                   }, item)
    end
  end

  def test_with_found_item_raises_error_when_archive_wrong
    frazzle = 'frazzle'
    frazzle_dazzle = File.join(frazzle, 'dazzle')
    frazzle_dazzle_foo = File.join(frazzle_dazzle, 'foo')
    frazzle_foo = File.join(frazzle, 'foo')

    FileUtils.mkdir_p(frazzle_dazzle)
    File.open(frazzle_foo,'wb') {|io| io.puts 'frazzle_foo'}
    File.open(frazzle_dazzle_foo,'wb') {|io| io.puts 'frazzle_dazzle_foo'}

    frazzle_dazzle_foo_hash = file_hash(frazzle_dazzle_foo)
    frazzle_dazzle_foo_mode = File.stat(frazzle_dazzle_foo).mode
    frazzle_foo_hash = file_hash(frazzle_foo)
    frazzle_foo_mode = File.stat(frazzle_foo).mode

    archive = encrypt_and_archive(frazzle)
    FileUtils.rm_rf(frazzle)
    assert(!File.exists?(frazzle))
    
    assert_raises RuntimeError do
      with_found_item("mazzle/frazzle/frazzle_dazzle/foo", archive) do |temp|
        print File.read(temp)
      end
    end
  end

  def test_with_found_item_raises_error_when_find_file_where_expect_directory
    frazzle = 'frazzle'
    frazzle_dazzle = File.join(frazzle, 'dazzle')
    frazzle_foo = File.join(frazzle, 'foo')
    frazzle_dazzle_foo = File.join(frazzle_dazzle, 'foo')

    FileUtils.mkdir_p(frazzle_dazzle)
    File.open(frazzle_foo,'wb') {|io| io.puts 'frazzle_foo'}
    File.open(frazzle_dazzle_foo,'wb') {|io| io.puts 'frazzle_dazzle_foo'}

    frazzle_dazzle_foo_hash = file_hash(frazzle_dazzle_foo)
    frazzle_dazzle_foo_mode = File.stat(frazzle_dazzle_foo).mode
    frazzle_foo_hash = file_hash(frazzle_foo)
    frazzle_foo_mode = File.stat(frazzle_foo).mode

    archive = encrypt_and_archive(frazzle)
    FileUtils.rm_rf(frazzle)
    assert(!File.exists?(frazzle))
    
    assert_raises RuntimeError do
      with_found_item("frazzle/frazzle_dazzle/foo/bar", archive) do |temp|
        print File.read(temp)
      end
    end
  end

  def test_with_found_item_raises_when_item_not_found
    frazzle = 'frazzle'
    frazzle_dazzle = File.join(frazzle, 'frazzle_dazzle')
    frazzle_foo = File.join(frazzle, 'foo')
    frazzle_dazzle_foo = File.join(frazzle_dazzle, 'foo')

    FileUtils.mkdir_p(frazzle_dazzle)
    File.open(frazzle_foo,'wb') {|io| io.puts 'frazzle_foo'}
    File.open(frazzle_dazzle_foo,'wb') {|io| io.puts 'dazzle_foo'}

    frazzle_dazzle_foo_hash = file_hash(frazzle_dazzle_foo)
    frazzle_dazzle_foo_mode = File.stat(frazzle_dazzle_foo).mode

    archive = encrypt_and_archive(frazzle)
    FileUtils.rm_rf(frazzle)
    assert(!File.exists?(frazzle))
    
    assert_raises RuntimeError do
      with_found_item(File.join(frazzle_dazzle, 'bar'), archive) do |item|
        true
      end
    end
  end

  def test_with_found_item_yields_item_when_correct_item_is_found
    frazzle = 'frazzle'
    frazzle_dazzle = File.join(frazzle, 'dazzle')
    frazzle_foo = File.join(frazzle, 'foo')
    frazzle_dazzle_foo = File.join(frazzle_dazzle, 'foo')

    FileUtils.mkdir_p(frazzle_dazzle)
    File.open(frazzle_foo,'wb') {|io| io.puts 'frazzle_foo'}
    File.open(frazzle_dazzle_foo,'wb') {|io| io.puts 'dazzle_foo'}

    frazzle_dazzle_foo_hash = file_hash(frazzle_dazzle_foo)
    frazzle_dazzle_foo_mode = File.stat(frazzle_dazzle_foo).mode

    archive = encrypt_and_archive(frazzle)
    FileUtils.rm_rf(frazzle)
    assert(!File.exists?(frazzle))
    
    with_found_item(frazzle_dazzle_foo, archive) do |item|
      assert_equal('file', item['type'])
      assert_equal('foo', item['name'])
      assert_equal(frazzle_dazzle_foo_mode, item['mode'])
      assert_match(/^[a-fA-F0-9]+$/, item['key'])
      assert_match(/^[a-fA-F0-9]+$/, item['address'])
    end
  end

  # TODO: how to recover a single file?
  # TODO: how to recover a single symlink?
  # TODO: how to recover a single directory?

end
