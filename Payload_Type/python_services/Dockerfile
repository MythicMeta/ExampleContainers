FROM itsafeaturemythic/mythic_python_go:latest

RUN pip3 install mythic_container==0.7.0rc9
RUN pip3 install --upgrade --force-reinstall grpcio>=1.78.0

WORKDIR /Mythic/

COPY [".", "."]

# fix Windows CRLF in .py files so shebangs and imports work
RUN find . -name '*.py' -exec sed -i 's/\r$//' {} +

RUN pip3 install -r requirements.txt

CMD ["python3", "main.py"]
