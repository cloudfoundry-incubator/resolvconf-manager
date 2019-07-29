# What does this do?

This repository acts as an abstraction layer between clients of
resolv.conf-managing apt packages (resolvconf and openresolv, for now) and the
stemcell. This abstraction ensures that those clients will not need to know
which manager is installed on a given stemcell since they will have a consistent
API for updating resolv.conf.

# Why create a utility for such a small thing?

Without this utility, bosh-dns and the bosh-agent would be tightly coupled to a
given stemcell's resolv.conf manager. This would mean that to change managers
we would need to update the agent. As we've already seen, which resolv.conf
manager is in favor is outside of our responsibility and thus we may need to
adapt to the community going forward.

# Why not have bosh-dns edit resolv.conf directly?

Currently, bosh-dns updates resolv.conf by utilizing the resolvconf package and
ultimately calling `resolvconf -u`. This ensures that clients of resolv.conf
are notified whenever the file changes. If bosh-dns edited the file directly,
those clients would not be notified. In general this would not be a problem,
but there is not a guarantee that bosh-dns will always come up first on any
given VM. Therefore, a client which should query bosh-dns first may have cached
a previous version of resolv.conf.
