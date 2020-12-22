class IofogctlAT<AT_VERSION> < Formula
  desc "Command line tool for deploying and administering ioFog platforms"
  homepage "https://github.com/eclipse-iofog/iofogctl"
  url "<URL>/<BUCKET>/<VERSION>/iofogctl.tar.gz"
  sha256 "<SHA256>"
  version "<VERSION>"

  depends_on "curl"
  depends_on "bash-completion"

  bottle :unneeded

  def install
    bin.install "iofogctl"
  end
end