----------------------------------------------------------------------

Name:    amber
Summary: One or more computers participating in a file backup ring.

Author:  Karrick McDermott <karrick@karrick.net>
Date:    2011-01-07

----------------------------------------------------------------------

DESCRIPTION

amber provides a user-friendly, safe and reliable data storage
facility for a group of people, controlled by that group of people.
No outsiders, no contracts, no service charges.

amber is "... Peer Data Storage of the People, by the People, and
for the People."

amber is "Your data in your cloud."

----------------------------------------------------------------------

USER CASES

I have immediate family that lives in Texas, and in Arizona, Florida,
and New York.  I live in New Jersey.  We all have computers and we all
have important data.  Wouldn't it be nice to be able to back up our
data to each other's computers, but also ensure privacy and
reliability?

My home network is always connected to the Internet, and could provide
a rally point for backing up data for my family, but where do I backup
my data?

I love some members of my family, but they just don't have the same
sense of humor that I have, and might find some of my files offensive.
Although I'm not doing anything illegal or immoral, I don't
necessarily want my family knowing what my data is, especially my
finances...  I'd like to make sure I can keep my data private, and I
really don't need to know what's in their files, as well.

----------------------------------------------------------------------

PROBLEM: Too many files to backup.

Over the years a typical computer user accumulates scores of important
files, and both hardware malfunctions and malware threaten the
integrity of those files.

Keeping a backup of a file or two is managable, but when the number of
files grows large, it becomes tedious and error-prone.  Usually even
the most dedicated users acknowledge they do not backup their files
often enough.

----------------------------------------------------------------------

PROBLEM: Multiple computers to manage backups for.

A typical user owns one or several computers, each with various files
that might be important to the user.  It is often too difficult to
backup a single computer, however, trying to manage the backups of
several computers can be a challenge many users don't even bother
attempting.

----------------------------------------------------------------------

WHY HAVE A BACKUP?

I've gone quite a while without lossing a hard drive or suffering data
loss, however, I have lost data before.  Upgrading my computer's
Operating System, I once lost my entire digital life.  Years later,
enduring a rough divorce, I lost the entire digital collection of
pictures of my children.

Most people go years without significant data loss, and you might be
lucky enough to never suffer as many people do.  However, if you are
concerned, and typical backup solutions worry or confuse you, then
perhaps this solution can meet your needs.

----------------------------------------------------------------------

HYPOTHETICAL DISCUSSION WITH A SKEPTIC (a.k.a. FAQ)

Q: I already have backup software.  Why would I need your solution?

A: Perhaps you do not.  How many copies of your backups do you keep?
How many types of media do you store your backups on?  Where do you
keep your backup media?  What would you do if your house suffered a
fire or flooding?

Q: I backup my data regularly, but my parents are not concerned or
don't know how.  I'd like to help them, but how?

A: Consider starting a amber network with your family members.  If
a few of your family members had computers always connected to the
Internet, they could serve as redundant backup servers for your
family.

Q: But I don't want my files visible to family members.

A: Neither do I, so I made sure each file is encrypted, and only you
can decrypt the contents of your files.  Even the filenames are
meaningless gibberish unless you unlock the files.

Q: That redundant backup server setup sounds like I need my nephew to
setup and maintain, not to mention the costs of the hardware.  How am
I supposed to afford that?

A: That's what the big companies want you to think, because their
proprietary solutions come at a high price.  Quite honestly, if you
can download Firefox on your computer, then you can start a amber
network.  I made it for my mom, so I'm sure you can use it too.

Q: Don't I need big iron computers, expensive high speed hard
drives, and that RAID thing?

A: Sure, if you want.  But those things only speed up access to your
data.  A very basic computer is all you really need to backup data.
Instead of high redundancy equipment at a single location, the
amber network was designed to use multiple inexpensive computers
at different geographic locations to provide the redundancy you need.

Q: But wouldn't it still be slow?

A: Downloading a file slowly is still faster than retyping it from
memory.  And don't ask me to rekey in the color code bytes for a
picture you accidentally erased of your last family reunion.

Q: Okay, you said it's easy to use.  You said I don't need a fancy
computer module or adapter to use.  What's the catch?

A: Somebody had to take those fundamental computer science problems
that we all face, and design a solution to that problem.  I have the
same problem, so I built a solution, and offer it to you also.

Q: That doesn't make sense.  Why not make a million dollars from it?

A: I tried that, and I suck at business skills.  I decided to stick to
my strengths and write good computer code.

Q: I don't buy it.  You're a scam, and you're trying to get my
secrets.

A: There are thousands of malware authors writting software to steal
your money already, and perhaps a few malware installed on your
computer.  They open up your webcam and look at you while you pay your
bills online.  Seriously.

Q: I still don't trust you.

A: Okay, well perhaps you could read some software reviews that you do
trust that discuss this software.  Or, just read the source code.  I'm
not hiding anything.

Q: Why are you so mad at me?

A: I'm not.  Most people never read this far into any sort of
documentation, and I thought it would be fun to pretend I'm talking to
a very skeptical person when writting this.

----------------------------------------------------------------------

SOLUTION

1) Download and install software.
2) Create user account.
3) Configure.
4) Relax.

----------------------------------------------------------------------

Distributed backup ring allows recovery of files from multiple places.
Each other place may be online or offline, so software must be able to
find a copy from a host that's online.

USE CASE: A typical user would have a single computer that is not
always online.

I'm leaning towards the user machine always maintaining the encrypted
version of the file corpus for efficiency purposes.  However, I may
end up with simply maintaining the encrypted version of the file long
enough to upload to the first amber server.  More consideration on
this point is required.

USE CASE: Many participating members would setup a file server acting
as a amber node.  It would be available online most of the time, but
the amber system should be able to tolerate occassional downtime from
one or more servers, typically with no impact on file recovery.

TODO: Look into how various distributed file systems manage
replication.

MECHANICS

Small ANSI C program watches a directory on source computer.  When
given the synchronize command, it inventories the directory contents
and compares to the account inventory list.

New and updated files are symetrically encrypted using the hash of the
file contents.  Files are then scheduled to be uploaded to replication
servers, which are merely other computers in the backup-ring.

Removed files?  Cannot remove from ring in case file removal was
mistaken, or in case another user in the ring has the same file.

TODO: Figure out garbage collection scheme.

NOTE: An updated file looks like a new file because of its hash.

PROTECTION OF FILE CONTENTS

Files are symetrically encrypted on source computer using hash of
original file content.  Encrypted files are stored in repository on
source computer, and distributed to several other computers, called
amber nodes.

A user account maintains an inventory file, which is a mapping of
encrypted files to original files.

----------------------------------------------------------------------

CASE 1: New account.

A blank inventory file is created.

----------------------------------------------------------------------

CASE 2: New files.

----------------------------------------------------------------------

CASE 3: Modified files.

----------------------------------------------------------------------

CASE 4: Updated files.

----------------------------------------------------------------------

CASE 5: Login using new computer.

----------------------------------------------------------------------

CASE 6: Delete account.

----------------------------------------------------------------------

CASE 7: New or modified file has same hash as existing file in same
account.

----------------------------------------------------------------------

CASE 8: New or modified file has same hash as existing file in
different account.

----------------------------------------------------------------------

CASE 9: Recovery of file from backup: backup available.

----------------------------------------------------------------------

CASE 10: Recovery of file from backup: backup not available.

TODO

----------------------------------------------------------------------

CASE 11: Server offline.

----------------------------------------------------------------------

CASE 12: Server online after offline period.

----------------------------------------------------------------------

NOTES

Servers are amber nodes configured to operate in a peer group to
backup and replicate files.  The collection of servers in a peer group
define the peer-group, however some servers may be occassionally
unavailable.

Workstations are clients configured to work with amber peer group.

There is nothing preventing a server from running a client to manage
files, however, this point is made for completeness.  The point is
there is a distinct separation from amber clients and amber
servers.  Both may coexist on a single computer when desired, although
it is not a good idea to run the server software daemon on a
workstation that is not always connected to the network.

User account information is managed by the software in a distributed
fassion as well.

Some form of versioning is required if the amber is able to restore
corrupted files.

----------------------------------------------------------------------

NOTE: Lots of the below is obsolete since adopting a git like scheme
for version management.

CLIENT SOFTWARE

Maintain an account inventory, which is a table of all files and
associated metadata in an account.

OBSOLETE: Monitor a directory hierarchy, looking for changes against
the account inventory.  Files which have a content hash not in the
inventory must be enqueued for backup.  Files in the inventory no
longer present in the hierarchy must be enqueued for distributed
obsolecence.

Provide ability to recover files from the distributed backup.

Maintain list of URLs to the peer group.  This list need not be
complete, but perhaps a cache of which servers have been used before.
Any server in the peer group can be used to talk to the backup
servers.

----------------------------------------------------------------------

SERVER SOFTWARE

Provide discovery service, similar to Bonjour or mdns.

Maintain list of URLs to the peer group.  Perhaps this list need not
be complete.

Provide ability for a server to join this server's peer group.
Likewise, provide abiltity to revoke a server's peer group
credentials.

Provide ability to enumerate other peer group servers.

Provide ability to create new user account.

Provide ability to delete existing user account.

Provide ability to download user account's encrypted inventory file.
(Recall that an inventory file is useless without user's symetric
password.  However, should a user be unable to recall their inventory
password, the user's remote file contents are effectively useless.)

----------------------------------------------------------------------

Now using directory files, and the user's inventory is merely the root
directory in their purview.  In this way, when a directory is found, a
directory file is created for it, encrypted, and stored.  This would
create lots more garbage, however, a user's inventory is kept to a
managable size.

A directory file is an inventory file for the contents of a directory.

Because updates to a directory file would percolate up the hierarchy,
would have to perform depth-first traversals of hierarchy.  A
directory would be modified if its directory file would be modified...

While attempting to find a way to traverse the directories, I tried
the below code:

----------------------------------------

Garbage collection:

How to do it with encryption?  Need the account file password to walk
the tree, so it will have to be a user command.  Perhaps garbage
collection is not needed because user is not rebasing their backups,
and in git, a rebase is what creates garbage.

Is there any reason for garbage collection?  Yes, to save hard drive space.

Adam and Brian both have links to the same file.  Adam updates his
version, but Brian still has a link to the original.  Later Brian
updates his file also, so there are now three files in the archive.
It is impossible to determine whether the original file may be deleted
because it is unknown whether any one else has a link to the original file.

With version control, Adam may want to restore the original version of
his file, so once again, performing garbage collection hinders
expected operation.

It would be expected that the corpus of files would continue growing
indefinitely.

Perhaps an amber server would update the accessed time of any files
associated with addresses for GET and PUT requests.  This way files
may be sorted on time since last accessed.

Another option is to allow removal of directory files, but this seems
futile.

I'm still not happy with the fact that the amount of hard drive space
required to archive a file set will always grow due to unlimited
amount of versioning.

----------------------------------------

TERMINOLOGY

I'm a bit annoyed at my terminology, and for lack of anything better,
it may be time to consider using git terms for amber objects.

From git terminology, types of objects in a git repository are:

blob -- data, usually a file

tree -- directory pointing to blobs and other trees

commit -- refers to a tree that represents the state of files at the time of a commit, and points to its parents.

refs -- pointers to commits of interest, not stored in repository

For backups, a user may consider the following:

Files, Directories, Backup event, Account

I'm not convinced that in amber I need both commits and refs, because
a union of the two seems well suited for tracking the user's data.

amber | git
file | blob
directory | tree
backup event | commit
account file | ref

I want to ensure I rely heavily on recursive algorithms where
possible.  Perhaps I'm wrong, but it seems that the data structures
used by git are only partially recursive.

account file -> backup event 2 -> directory -> directory -> file
                               \> backup event 1 -> directory -> file

creation of backup event also updates user's account file to point to
most recent backup event.

================

Q: Can user have more than one account?

A: Not directly.  It's best if each user has a single account.

Q: How can a single user back up multiple machines or repository roots?

A: Maybe the account file has a collection of repositories, and the
   user selects which repository to mirror.

Q: Could you show me an example?

A: Sure.

$ cd ~ ; amber clone karrick@karrick.org:foo
foo directory is created, along with foo/.amber/
<password> entered to store cached version of account file
foo directory is populated with most recent backup event

$ cd foo ; amber info
karrick@karrick.org:foo

$ amber status
...

$ cd ~ ; amber clone karrick@karrick.org:bar

account file: [repos1, repos2, ...]

repos: points to most recent backup

backup: points to directory blob

Data is encrypted, so server does not need to be ssl.

Q: I have an amber account.  How do I backup an existing directory?

A: Demonstration:

$ cd foo ; amber init karrick@karrick.org

This 

================

How do I reconsile existing HEAD and HEAD from another computer?

If the user's account file on one system is updated, before the
accounts file is uploaded from a different computer, how to make sure
changes are merged?

It seems that I may have one account file, but many repos.
================

queue-manager can be used as the server.  It doesn't have TLS, but
I've have already established I don't need it.  I *do* need to make
certain the client does not overwrite data it should not.

================

Perhaps the same executable can be used as an amber server also.  When
given the --server argument, it binds to a network port and awaits
file requests.  It would answer GET and PUT requests, but PUT requests
would only be honored if (1) the address was not yet present in the
archive, and (2) the hash of the data matched the address.

If the request was a GET request, then check whether the file is in
the archive.  If so, send data stream back to requester.

If the file is not in the archive, tell the requester the file is not
present, then check to ensure the address was not listed on its list
of addresses it is searching for, preventing recursive loops in the
network.

If the file is not in the archive, and the address is not already
being searched for, then append the address to the search list.

On a timer, for all addresses in the search list, request the file
from any configured peers.  Once a searched for file is retrieved and
validated, save it in the archive, and no further action is required.  

================

Might want to consider using a longer hash to identify files.  SHA1
has a smaller bit space and may have collissions in the encrypted file
space.  I'm not concerned with plain text collissions, because the
hash of the plain text is only used for the symmetric cipher key, not
to give the data its address.

You might also consider storing the hash type, and using the
appropriate hash function for verification.  This would allow the
flexibility of changing which hash is used without having to
completely reprocess all files in the archive.  You may consider doing
the same with the encryption algorithm used.  In this way, one would
use the appropriate decryption algorithm for each file, yet be able to
use a newer algorithm for encrypting new files.

================

You may want to also consider a compression option.  Files could be
compressed prior to encryption and archival.  If compressed, their
item hash would indicate the compression algorithm used, so automatic
decompression would be performed when restoring the file's contents.

================

2013-08-21

URL:

user/{user-hash}/{resource-hash}
cache/{resource-hash}

To perform *any* operation on a user resource, must sign with user's private key.

To GET, sign the url.
To PUT, sign the content.
To DELETE, sign the url.

When POSTing, the content appends to existing content, rather than overwriting it.
To support mail box, when POSTing to a mail box, no signature is required.

POST user/{recipient-hash}/MAIL/{resource-hash}

For the user to check email, they send a signed GET to:

GET user/{user-hash}

If the signature is correct, they receive content back, including a list of newly posted
content in their inbox:

user/{recipient-hash}/MAIL/{resource-hash}
user/{recipient-hash}/MAIL/{resource-hash}
user/{recipient-hash}/MAIL/{resource-hash}

To retrieve an item, they send a signed GET on each of those resources.

================

To support new hash and encryption algorithms, each account and
resource is tagged with the hash and encryption algorithm used.

sha512/aes/user/{recipient-sha512-hash}/{resource-sha512-hash}

And a community resource:

sha512/aes/{resource-sha512-hash}

This does require a laborous process to change a user's preferred hash
and encryption algorithm, but only if all old resources are
upgraded. Existing data may remain unaltered, and be available in
perpetuity.

If a security flaw is discovered in a particular hash type or
encrpytion algorithm, a user may choose to upgrade data to a newer
hash or algorithm, but this process will likely take a prelonged
period of time to complete.

1. walk tree, downloading items
2. decrypt item on local machine with old hash and algorithm
3. encrypt item with new hash and/or algorithm
4. upload item
5. When tree walk is complete, walk old tree and perform a depth-first
delete of all resource items.

It would be good to periodically checkpoint by creating a directory
with items, some of which have been converted, some of which have yet
to be converted.

A weakness of putting the hash and encryption algorithm in the url is
that once a security flaw is discovered, it immediately allows probing
to try to unlock secrets.

================

There is plausible deniability when resources have a community
resource label. That is, anyone can upload and anyone can download
them. The downloads will hopefully be illegible without the key used
to encrypt the resource. The uploads will not pass early if they have
already been uploaded. This prevents probing for resources on a
server. The server will still verify the hash on upload, and yield a
pass or fail based on the hash being correct.

================

SAFE METHODS:
	GET
	HEAD
IDEMPOTENT METHODS:
	GET
	HEAD
	PUT
	DELETE
	OPTIONS
	TRACE

================

GET MAIL: Use If-Modified-Since field to allow partial
retrieval. Other similar fields, such as If-Range, may allow other
optimizations.

================

HEAD is identical to GET, other than no content body is transmitted
back. This can be used to check the size of a resource prior to
getting it.

================

PUT: after putting a resource, must respond with 201 if resource
created. If resource modified, server should respond with either 200
or 204, but due to privacy concerns, Amber will return 201
also. Standards allow this behavior.

