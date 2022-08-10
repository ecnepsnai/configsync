Name:           configsync
Version:        %{_version}
Release:        1
Summary:        Configuration synchronization software
License:        MIT
Source0:        %{name}-%{version}.tar.gz
BuildRequires:  systemd-rpm-macros
Provides:       %{name} = %{version}
Prefix:         %{_sbindir}

%description
A command line tool to sync system configuration files with a git repo.

%global debug_package %{nil}

%prep
%autosetup

%build
cd cmd/configsync
CGO_ENABLED=0 GOAMD64=v2 go build -buildmode=exe -trimpath -ldflags="-s -w -X 'main.Version=%{version}' -X 'main.BuildDate=%{_date}'" -v -o configsync
./configsync -v

%install
install -Dpm 0755 cmd/configsync/configsync %{buildroot}/%{_sbindir}/configsync

%check
CGO_ENABLED=0 GOAMD64=v2 go build -v ./...
CGO_ENABLED=0 GOAMD64=v2 go test -v ./...

%files
%{_sbindir}/configsync
