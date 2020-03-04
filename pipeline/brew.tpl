class IofogctlAT<AT_VERSION> < Formula
  desc "Command line tool for deploying and administering ioFog platforms"
  homepage "https://github.com/eclipse-iofog/iofogctl"
  url "<URL>/<REL_BUCKET>/<REL_VERSION>/iofogctl.tar.gz"
  sha256 "<REL_SHA256>"
  version "<REL_VERSION>"
  devel do
    url "<URL>/<DEV_BUCKET>/<DEV_VERSION>/iofogctl.tar.gz"
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