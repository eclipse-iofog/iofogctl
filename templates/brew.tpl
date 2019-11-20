class IofogctlAT<AT_VERSION> < Formula
  desc "Command line tool for deploying and administering ioFog platforms"
  homepage "https://github.com/eclipse-iofog/iofogctl"
  url "http://edgeworx.io/downloads/iofogctl/rel/<REL_VERSION>.tar.gz"
  sha256 "<REL_SHA256>"
  version "<REL_VERSION>"
  devel do
    url "http://edgeworx.io/downloads/iofogctl/dev/<DEV_VERSION>.tar.gz"
    sha256 "<DEV_SHA256>"
    version "<DEV_VERSION>"
  end

  depends_on "curl"
  depends_on "bash-completion"

  bottle :unneeded

  def install
    bin.install "iofogctl"
  end
end