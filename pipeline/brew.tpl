class IofogctlAT<AT_VERSION> < Formula
  desc "Command line tool for deploying and administering ioFog platforms"
  homepage "https://github.com/eclipse-iofog/iofogctl"

  version "<VERSION>"

  depends_on "curl"
  depends_on "bash-completion"

  bottle :unneeded
  on_macos do
    if Hardware::CPU.arm?
      url "<URL>/<BUCKET>/<VERSION>/iofogctl_arm64.tar.gz"
      sha256 "<SHA256>"

      def install
        bin.install "edgectl"
      end
    end
    if Hardware::CPU.intel?
      url "<URL>/<BUCKET>/<VERSION>/iofogctl_amd64.tar.gz"
      sha256 "<SHA256>"

      def install
        bin.install "edgectl"
      end
    end
  end
end